package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/rs/xid"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

//go:embed *.html
var tpls embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("Failed to start server", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env file is present.
	godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return errors.New("missing JWT_SECRET environment variable")
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		return errors.New("missing HTTP_PORT environment variable")
	}

	jwt := jwtauth.New("HS256", []byte(secret), nil)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestSize(32000))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(jwtauth.Verifier(jwt))

	t, err := template.New("tpls").Funcs(sprig.FuncMap()).ParseFS(tpls, "*.html")
	if err != nil {
		return fmt.Errorf("parse template fs: %w", err)
	}

	room := newRoom()
	// Seeding random messages in room.
	for i := 0; i < 1000; i++ {
		msg, _ := newMessage(
			newUser(),
			faker.Sentence(options.WithGenerateUniqueValues(false)),
		)
		room.addMessage(msg)
	}

	l := newLimiters()

	// Protected routes.
	r.Group(func(r chi.Router) {
		r.Use(protected)

		r.Get("/", index(t, room))
		r.Handle("/chatroom", websocket.Handler(chat(t, room, l)))
	})

	r.Get("/login", login(jwt))

	server := &http.Server{
		Addr:         "localhost:" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	slog.Info("Running server...", "addr", "http://"+server.Addr)
	return server.ListenAndServe()
}

func index(t *template.Template, room *room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.ExecuteTemplate(w, "index.html", map[string]any{
			"User":     user,
			"Messages": room.listMessages(),
			"NumUsers": room.numUsers(),
			"Disabled": false,
		})
	}
}

func login(auth *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		if err == nil && token != nil && jwt.Validate(token) == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Create a fake user and use it as claim to encode a jwt token.
		_, t, err := auth.Encode(map[string]any{
			"user": newUser(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    t,
			Expires:  time.Now().Add(1 * time.Hour),
			Secure:   false,
			HttpOnly: false,
			Path:     "/",
		})

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

type data struct {
	Message string            `json:"chat_message"`
	Headers map[string]string `json:"HEADERS"`
}

func chat(t *template.Template, r *room, l *limiters) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		defer ws.Close()

		var b bytes.Buffer
		// Retrieve user from context.
		ctx := ws.Request().Context()
		user := userFromContext(ctx)
		added := r.addClient(user, ws)
		logger := slog.Default().With("user.id", user.ID)
		// Remove client from room when user disconnects.
		defer func() {
			if r.removeClient(user.ID, ws) {
				// If user is fully disconnected, remove limiter.
				l.remove(user)

				b.Reset()

				// Update number of user online for all users.
				if err := t.ExecuteTemplate(&b, "online", map[string]any{
					"NumUsers": r.numUsers(),
				}); err != nil {
					logger.ErrorContext(ctx, "compile online template", "err", err)
					return
				}
				r.broadcast(b.String())
			}
		}()

		// If added, update number of user online for all users.
		if added {
			if err := t.ExecuteTemplate(&b, "online", map[string]any{
				"NumUsers": r.numUsers(),
			}); err != nil {
				logger.ErrorContext(ctx, "compile online template", "err", err)
				return
			}
			r.broadcast(b.String())
		}

		limiter := l.add(user, 5*time.Second, 3)

		for {
			b.Reset()

			var d data
			if err := websocket.JSON.Receive(ws, &d); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				logger.ErrorContext(ctx, "receive message", "err", err)

				// Inform user something went wrong.
				if err := t.ExecuteTemplate(&b, "error", map[string]any{"Error": "could not read your message"}); err != nil {
					logger.ErrorContext(ctx, "compile error template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send error", "err", err)
				}

				continue
			}

			// Rate limit to prevent abuse.
			if !limiter.Allow() {
				// Inform the current user to slow down.
				if err := t.ExecuteTemplate(&b, "error", map[string]any{"Error": "why so fast?"}); err != nil {
					logger.ErrorContext(ctx, "compile error template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send error", "err", err)
					continue
				}

				b.Reset()

				// Disable the form until limiter allows.
				if err := t.ExecuteTemplate(&b, "form", map[string]any{
					"Disabled": true,
				}); err != nil {
					logger.ErrorContext(ctx, "compile form template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send form", "err", err)
					continue
				}

				if err := limiter.Wait(ctx); err != nil {
					logger.ErrorContext(ctx, "limiter wait", "err", err)
					continue
				}

				b.Reset()

				// Re-enable the form.
				if err := t.ExecuteTemplate(&b, "form", map[string]any{
					"Disabled": false,
				}); err != nil {
					logger.ErrorContext(ctx, "compile form template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send form", "err", err)
					continue
				}

				// Clear the error for the current user.
				if err := t.ExecuteTemplate(&b, "error", map[string]any{"Error": ""}); err != nil {
					logger.ErrorContext(ctx, "compile error template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send form", "err", err)
				}

				continue
			}

			// Create and add the message to the room.
			msg, err := newMessage(user, d.Message)
			if err != nil {
				// Send back an error if we could not create message.
				// Could be a validation error.
				if err := t.ExecuteTemplate(&b, "error", map[string]any{"Error": err.Error()}); err != nil {
					logger.ErrorContext(ctx, "compile error template", "err", err)
					continue
				}
				if err := websocket.Message.Send(ws, b.String()); err != nil {
					logger.ErrorContext(ctx, "send error", "err", err)
				}

				continue
			}
			r.addMessage(msg)

			// Broadcast message to all clients including the current user.
			if err := t.ExecuteTemplate(&b, "message", msg); err != nil {
				logger.ErrorContext(ctx, "compile message template", "err", err)
				continue
			}
			r.broadcast(`<div hx-swap-oob="beforebegin:#messages>li:last-child">` + b.String() + `</div>`)

			b.Reset()

			// Reset the form for the current user.
			if err := t.ExecuteTemplate(&b, "form", map[string]any{
				"Disabled": false,
			}); err != nil {
				logger.ErrorContext(ctx, "compile form template", "err", err)
				continue
			}
			if err := websocket.Message.Send(ws, b.String()); err != nil {
				logger.ErrorContext(ctx, "send form", "err", err)
			}

			b.Reset()

			// Clear the error for the current user.
			if err := t.ExecuteTemplate(&b, "error", map[string]any{"Error": ""}); err != nil {
				logger.ErrorContext(ctx, "compile error template", "err", err)
				continue
			}
			if err := websocket.Message.Send(ws, b.String()); err != nil {
				logger.ErrorContext(ctx, "send form", "err", err)
			}
		}
	}
}

func protected(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil || jwt.Validate(token) != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Retrieve the user from the claims and add it to the request context.
		// If the user ID is invalid, we attempt login again.
		// This could lead to an infinite loop if a user has a newer claim format.
		u := claims["user"].(map[string]any)
		id, err := xid.FromString(u["ID"].(string))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := addUserToContext(r.Context(), &user{
			ID:   id,
			Name: u["Name"].(string),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
