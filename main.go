package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mgjules/chat-demo/chat"
	"github.com/mgjules/chat-demo/templates"
	"github.com/mgjules/chat-demo/user"
	"github.com/rs/xid"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

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

	room := chat.NewRoom()
	// Seeding random messages in room.
	for i := 0; i < 1000; i++ {
		msg, _ := chat.NewMessage(
			user.New(),
			faker.Sentence(options.WithGenerateUniqueValues(false)),
		)
		room.AddMessage(msg)
	}

	l := newLimiters()

	// Protected routes.
	r.Group(func(r chi.Router) {
		r.Use(protected)

		r.Get("/", index(room))
		r.Handle("/chatroom", websocket.Handler(chatroom(room, l)))
	})

	r.Get("/login", login(jwt))

	server := &http.Server{
		Addr:         ":" + port,
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
			"user": user.New(),
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

		ctx := user.AddToContext(r.Context(), &user.User{
			ID:   id,
			Name: u["Name"].(string),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func index(room *chat.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := user.FromContext(r.Context())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.Page(user, room, false, "").Render(r.Context(), w); err != nil {
			slog.ErrorContext(r.Context(), "render index template", "err", err, "user.id", user.ID)
			w.Write([]byte("failed to render index template"))
		}
	}
}

type data struct {
	Message string            `json:"chat_message"`
	Headers map[string]string `json:"HEADERS"`
}

func chatroom(r *chat.Room, l *limiters) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		ws.MaxPayloadBytes = 2 << 10 // 2KB
		defer ws.Close()

		// Retrieve user from context.
		ctx := ws.Request().Context()
		u := user.FromContext(ctx)
		added := r.AddClient(u, ws)
		logger := slog.Default().With("user.id", u.ID)
		// Remove client from room when user disconnects.
		defer func() {
			if r.RemoveClient(u.ID, ws) {
				// If user is fully disconnected, remove limiter.
				l.remove(u)

				// Update number of user online for all users.
				if err := templates.ChatHeaderNumUsers(r.NumUsers()).Render(ctx, r); err != nil {
					logger.ErrorContext(ctx, "render online template", "err", err)
				}
			}
		}()

		// If added, update number of user online for all users.
		if added {
			if err := templates.ChatHeaderNumUsers(r.NumUsers()).Render(ctx, r); err != nil {
				logger.ErrorContext(ctx, "render online template", "err", err)
				return
			}
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
				if err := templates.ChatError("could not read your message").Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}

				continue
			}

			// Rate limit to prevent abuse.
			if !limiter.Allow() {
				// Inform the current user to slow down and
				// disable the form until limiter allows.
				if err := templates.ChatForm(true).Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render form template", "err", err)
					break
				}
				if err := templates.ChatError("please slow down").Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}

				// Wait until user is no more rate-limited
				if err := limiter.Wait(ctx); err != nil {
					logger.ErrorContext(ctx, "limiter wait", "err", err)
					continue
				}

				// Re-enable the form.
				// Clear the error for the current user.
				if err := templates.ChatForm(false).Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render form template", "err", err)
					break
				}
				if err := templates.ChatError("").Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}

				continue
			}

			// Create and add the message to the room.
			msg, err := chat.NewMessage(u, d.Message)
			if err != nil {
				// Send back an error if we could not create message.
				// Could be a validation error.
				if err := templates.ChatError(err.Error()).Render(ctx, ws); err != nil {
					logger.ErrorContext(ctx, "render error template", "err", err)
					break
				}

				continue
			}
			r.AddMessage(msg)

			// Broadcast personalized message to all clients including the current user.
			r.IterateClients(func(u *user.User, conn *websocket.Conn) error {
				if err := templates.ChatMessageWrapped(u, msg).Render(ctx, conn); err != nil {
					return fmt.Errorf("render message template: %w", err)
				}

				return nil
			})

			// Reset the form and clear the error for the current user.
			if err := templates.ChatForm(false).Render(ctx, ws); err != nil {
				logger.ErrorContext(ctx, "render form template", "err", err)
				break
			}
			if err := templates.ChatError("").Render(ctx, ws); err != nil {
				logger.ErrorContext(ctx, "render error template", "err", err)
				break
			}
		}
	}
}
