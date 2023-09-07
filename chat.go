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
	user *user
	conn *websocket.Conn
}

func newClient(u *user, c *websocket.Conn) *client {
	return &client{
		user: u,
		conn: c,
	}
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

func (r *room) addClient(c *client) {
	if c == nil {
		return
	}

	r.mu.Lock()
	r.clients[c.user.ID.String()] = c
	r.mu.Unlock()
}

func (r *room) removeClient(id xid.ID) {
	r.mu.Lock()
	delete(r.clients, id.String())
	r.mu.Unlock()
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
	for _, c := range r.clients {
		if err := websocket.Message.Send(c.conn, string(b)); err != nil {
			slog.WarnContext(c.conn.Request().Context(), "send message", "err", "user.id", c.user.ID)
		}
	}
}
