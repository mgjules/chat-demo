package chat

import (
	"container/ring"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/enescakir/emoji"
	"github.com/mgjules/chat-demo/user"
	"github.com/rs/xid"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

const maxMessageSize = 256

// Message represents a single chat message.
type Message struct {
	User    *user.User
	Content string
	Time    time.Time
}

// NewMessage creates a new Message.
func NewMessage(u *user.User, content string) (*Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	rc := []rune(content)
	if len(rc) > maxMessageSize {
		content = string(rc[:maxMessageSize]) + "..."
	}

	content = goaway.Censor(emoji.Parse(content))

	return &Message{
		User:    u,
		Content: content,
		Time:    time.Now().UTC(),
	}, nil
}

// Client represents the relationship between a user and websocket connections.
type Client struct {
	user  *user.User
	conns map[*websocket.Conn]struct{}
}

// Room holds the state of a single chat room.
type Room struct {
	muClients sync.RWMutex
	clients   map[string]*Client

	muMessages sync.RWMutex
	messages   *ring.Ring
}

// NewRoom creates a new Room.
func NewRoom() *Room {
	return &Room{
		clients:  make(map[string]*Client),
		messages: ring.New(100),
	}
}

// AddClient adds a websocket connection to a user as a client
// If the user does not already have a connection, thus no client
// it will be created and the method will return true.
func (r *Room) AddClient(u *user.User, ws *websocket.Conn) bool {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	id := u.ID.String()
	var added bool
	if _, found := r.clients[id]; !found {
		r.clients[id] = &Client{
			user:  u,
			conns: make(map[*websocket.Conn]struct{}),
		}
		added = true
	}
	r.clients[id].conns[ws] = struct{}{}

	return added
}

// RemoveClient removes a websocket connection from a user.
// If the user does not have any websocket connection, its client will be removed
// and the method will return true.
func (r *Room) RemoveClient(id xid.ID, ws *websocket.Conn) bool {
	r.muClients.Lock()
	defer r.muClients.Unlock()
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

// NumUsers return the current number of users as clients.
func (r *Room) NumUsers() uint64 {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	return uint64(len(r.clients))
}

// AddMessage adds a new chat message.
func (r *Room) AddMessage(m *Message) {
	r.muMessages.Lock()
	r.messages.Value = m
	r.messages = r.messages.Next()
	r.muMessages.Unlock()
}

// Messages returns the list of messages.
func (r *Room) Messages() []*Message {
	r.muMessages.RLock()
	defer r.muMessages.RUnlock()

	messages := make([]*Message, 0)
	r.messages.Do(func(m any) {
		messages = append(messages, m.(*Message))
	})

	return messages
}

// Write implements the io.Writer interface.
func (r *Room) Write(p []byte) (int, error) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	writers := make([]io.Writer, 0)
	for _, c := range r.clients {
		for conn := range c.conns {
			writers = append(writers, conn)
		}
	}

	return io.MultiWriter(writers...).Write(p)
}

// IterateClients executes a function fn
// (e.g. a custom send mechanism or personalized messages per client) for all the clients.
func (r *Room) IterateClients(fn func(u *user.User, conn *websocket.Conn) error) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	for _, c := range r.clients {
		for conn := range c.conns {
			if err := fn(c.user, conn); err != nil {
				slog.WarnContext(conn.Request().Context(), "send message", "err", "user.id", c.user.ID)
			}
		}
	}
}
