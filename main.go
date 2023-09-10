package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/flosch/pongo2/v6"
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

	ts := pongo2.NewSet("tpls", pongo2.NewFSLoader(tpls))
	t, err := ts.FromFile("index.html")
	if err != nil {
		return fmt.Errorf("failed to load index.html template: %w", err)
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

func index(t *pongo2.Template, room *room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteWriter(pongo2.Context{
			"user":      user,
			"messages":  room.listMessages(),
			"num_users": room.numUsers(),
			"disabled":  false,
		}, w); err != nil {
			slog.ErrorContext(r.Context(), "render index template", "err", err, "user.id", user.ID)
			w.Write([]byte("failed to render index template"))
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

type data struct {
	Message string            `json:"chat_message"`
	Headers map[string]string `json:"HEADERS"`
}

func chat(t *pongo2.Template, r *room, l *limiters) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		defer ws.Close()

		// Retrieve user from context.
		ctx := ws.Request().Context()
		u := userFromContext(ctx)
		added := r.addClient(u, ws)
		logger := slog.Default().With("user.id", u.ID)
		// Remove client from room when user disconnects.
		defer func() {
			if r.removeClient(u.ID, ws) {
				// If user is fully disconnected, remove limiter.
				l.remove(u)

				// Update number of user online for all users.
				res, err := t.ExecuteBlocks(pongo2.Context{
					"num_users": r.numUsers(),
				}, []string{"online"})
				if err != nil {
					logger.ErrorContext(ctx, "render online template", "err", err)
					return
				}
				r.broadcast(res["online"])
			}
		}()

		// If added, update number of user online for all users.
		if added {
			res, err := t.ExecuteBlocks(pongo2.Context{
				"num_users": r.numUsers(),
			}, []string{"online"})
			if err != nil {
				logger.ErrorContext(ctx, "render online template", "err", err)
				return
			}
			r.broadcast(res["online"])
		}

		limiter := l.add(u, 5*time.Second, 3)

		// Receiving and processing client requests.
		for {
			var d data
			if err := websocket.JSON.Receive(ws, &d); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				logger.ErrorContext(ctx, "receive message", "err", err)

				// Inform user something went wrong.
				res, err := t.ExecuteBlocks(pongo2.Context{
					"error": "could not read your message",
				}, []string{"error"})
				if err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}
				if err := websocket.Message.Send(ws, res["error"]); err != nil {
					logger.ErrorContext(ctx, "send message", "err", err)
				}

				continue
			}

			// Rate limit to prevent abuse.
			if !limiter.Allow() {
				// Inform the current user to slow down and
				// disable the form until limiter allows.
				res, err := t.ExecuteBlocks(pongo2.Context{
					"error":    "why so fast?",
					"disabled": true,
				}, []string{"error", "form"})
				if err != nil {
					logger.ErrorContext(ctx, "render error and form templates", "err", err)
					break
				}
				if err := websocket.Message.Send(ws, res["error"]+res["form"]); err != nil {
					logger.ErrorContext(ctx, "send message", "err", err)
				}

				if err := limiter.Wait(ctx); err != nil {
					logger.ErrorContext(ctx, "limiter wait", "err", err)
					continue
				}

				// Re-enable the form.
				// Clear the error for the current user.
				res, err = t.ExecuteBlocks(pongo2.Context{
					"error":    "",
					"disabled": false,
				}, []string{"error", "form"})
				if err != nil {
					logger.ErrorContext(ctx, "render error and form templates", "err", err)
					break
				}
				if err := websocket.Message.Send(ws, res["error"]+res["form"]); err != nil {
					logger.ErrorContext(ctx, "send message", "err", err)
				}

				continue
			}

			// Create and add the message to the room.
			msg, err := newMessage(u, d.Message)
			if err != nil {
				// Send back an error if we could not create message.
				// Could be a validation error.
				res, err := t.ExecuteBlocks(pongo2.Context{
					"error": err.Error(),
				}, []string{"error"})
				if err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}
				if err := websocket.Message.Send(ws, res["error"]); err != nil {
					logger.ErrorContext(ctx, "send message", "err", err)
				}

				continue
			}
			r.addMessage(msg)

			// Broadcast message to all clients including the current user.
			r.broadcastCustom(func(u *user, conn *websocket.Conn) error {
				// Broadcast message to all clients including the current user.
				res, err := t.ExecuteBlocks(pongo2.Context{
					"user": u,
					"msg":  msg,
				}, []string{"message"})
				if err != nil {
					return fmt.Errorf("render message template: %w", err)
				}
				if err := websocket.Message.Send(
					conn,
					`<div hx-swap-oob="beforebegin:#messages>li:last-child">`+res["message"]+`</div>`,
				); err != nil {
					return fmt.Errorf("send message: %w", err)
				}

				return nil
			})

			// Reset the form and clear the error for the current user.
			res, err := t.ExecuteBlocks(pongo2.Context{
				"error":    "",
				"disabled": false,
			}, []string{"error", "form"})
			if err != nil {
				logger.ErrorContext(ctx, "render error and form templates", "err", err)
				break
			}
			if err := websocket.Message.Send(ws, res["error"]+res["form"]); err != nil {
				logger.ErrorContext(ctx, "send message", "err", err)
			}
		}
	}
}
