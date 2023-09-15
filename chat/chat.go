package chat

import (
	"container/ring"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/enescakir/emoji"
	"github.com/mgjules/chat-demo/user"
	"github.com/rs/xid"
	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
	"golang.org/x/sync/semaphore"
)

const (
	maxMessageSize uint16 = 256
	maxSendWorker  uint16 = 1000
	maxClients     uint16 = 1000
)

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
	if len(rc) > int(maxMessageSize) {
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
	user *user.User
	conn *websocket.Conn
}

// Room holds the state of a single chat room.
type Room struct {
	muClients sync.RWMutex
	clients   map[string]*Client

	muMessages sync.RWMutex
	messages   *ring.Ring
	sem        *semaphore.Weighted
}

// NewRoom creates a new Room.
func NewRoom() *Room {
	return &Room{
		clients:  make(map[string]*Client),
		messages: ring.New(100),
		sem:      semaphore.NewWeighted(int64(maxSendWorker)),
	}
}

// AddClient adds a client along with its websocket connection.
func (r *Room) AddClient(u *user.User, ws *websocket.Conn) error {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	id := u.ID.String()
	if _, found := r.clients[id]; found {
		return errors.New("you can only have one instance of the chat")
	}

	if len(r.clients) >= int(maxClients) {
		return errors.New("room is full. please retry later")
	}

	r.clients[id] = &Client{
		user: u,
		conn: ws,
	}

	return nil
}

// RemoveClient removes a client.
func (r *Room) RemoveClient(id xid.ID) bool {
	r.muClients.Lock()
	defer r.muClients.Unlock()
	_, found := r.clients[id.String()]
	if !found {
		return false
	}

	delete(r.clients, id.String())

	return true
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

	r.IterateClients(func(u *user.User, conn *websocket.Conn) error {
		if _, err := conn.Write(p); err != nil {
			return fmt.Errorf("write: %w", err)
		}

		return nil
	})

	return len(p), nil
}

// IterateClients executes a function fn
// (e.g. a custom send mechanism or personalized messages per client) for all the clients.
func (r *Room) IterateClients(fn func(u *user.User, conn *websocket.Conn) error) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()

	var wg sync.WaitGroup
	for _, c := range r.clients {
		if err := r.sem.Acquire(c.conn.Request().Context(), 1); err != nil {
			slog.WarnContext(c.conn.Request().Context(), "acquire lock", "err", err, "user.id", c.user.ID)
			continue
		}

		wg.Add(1)
		go func(c *Client) {
			defer func() {
				r.sem.Release(1)
				wg.Done()
			}()

			if err := fn(c.user, c.conn); err != nil {
				slog.WarnContext(c.conn.Request().Context(), "send message", "err", "user.id", c.user.ID)
			}
		}(c)
	}

	wg.Wait()
}
