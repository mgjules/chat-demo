package main

import (
	"container/ring"
	"errors"
	"strings"
	"sync"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/enescakir/emoji"
	"github.com/rs/xid"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

const maxMessageSize = 256

type message struct {
	User    *user
	Content string
	Time    time.Time
}

func newMessage(u *user, content string) (*message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	rc := []rune(content)
	if len(rc) > maxMessageSize {
		content = string(rc[:maxMessageSize]) + "..."
	}

	content = goaway.Censor(emoji.Parse(content))

	return &message{
		User:    u,
		Content: content,
		Time:    time.Now().UTC(),
	}, nil
}

type client struct {
	user  *user
	conns map[*websocket.Conn]struct{}
}

type room struct {
	mu       sync.RWMutex
	clients  map[string]*client
	messages *ring.Ring
}

func newRoom() *room {
	return &room{
		clients:  make(map[string]*client),
		messages: ring.New(100),
	}
}

func (r *room) addClient(u *user, ws *websocket.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := u.ID.String()
	var added bool
	if _, found := r.clients[id]; !found {
		r.clients[id] = &client{
			user:  u,
			conns: make(map[*websocket.Conn]struct{}),
		}
		added = true
	}
	r.clients[id].conns[ws] = struct{}{}

	return added
}

func (r *room) removeClient(id xid.ID, ws *websocket.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, found := r.clients[id.String()]
	if !found {
		return false
	}

	delete(r.clients[id.String()].conns, ws)
	if len(r.clients[id.String()].conns) == 0 {
		delete(r.clients, id.String())
		return true
	}

	return false
}

func (r *room) numUsers() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.clients)
}

func (r *room) addMessage(m *message) {
	r.mu.Lock()
	r.messages.Value = m
	r.messages = r.messages.Next()
	r.mu.Unlock()
}

func (r *room) listMessages() []*message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages := make([]*message, 0)
	r.messages.Do(func(m any) {
		messages = append(messages, m.(*message))
	})

	return messages
}

func (r *room) broadcast(b string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.clients {
		for conn := range c.conns {
			if err := websocket.Message.Send(conn, b); err != nil {
				slog.WarnContext(conn.Request().Context(), "send message", "err", "user.id", c.user.ID)
			}
		}
	}
}

func (r *room) broadcastCustom(fn func(u *user, conn *websocket.Conn) error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.clients {
		for conn := range c.conns {
			if err := fn(c.user, conn); err != nil {
				slog.WarnContext(conn.Request().Context(), "send message", "err", "user.id", c.user.ID)
			}
		}
	}
}
