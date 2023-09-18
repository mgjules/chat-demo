package chat

import (
	"container/ring"
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

// List of chat errors.
var (
	ErrLoading         = NewError(ErrorSeverityWarning, true, "loading...")
	ErrUnknown         = NewError(ErrorSeverityError, false, "unknown error")
	ErrRateLimited     = NewError(ErrorSeverityWarning, false, "please slow down")
	ErrMessageEmpty    = NewError(ErrorSeverityError, false, "message content cannot be empty")
	ErrExistingSession = NewError(ErrorSeverityError, true, "you already have a running session for this room")
	ErrRoomFull        = NewError(ErrorSeverityError, true, "the room is full")
)

// ErrorSeverity is the severity of an error.
type ErrorSeverity uint8

// List of chat errors.
const (
	ErrorSeverityWarning ErrorSeverity = iota
	ErrorSeverityError
)

// Error represents a chat error.
type Error struct {
	err      string
	severity ErrorSeverity
	global   bool
}

// NewError creates a new Error.
func NewError(s ErrorSeverity, global bool, err string) Error {
	return Error{severity: s, global: global, err: err}
}

// IsError checks if the error of severity error.
func (e Error) IsError() bool { return e.severity == ErrorSeverityError }

// IsWarning checks if the error of severity warning.
func (e Error) IsWarning() bool { return e.severity == ErrorSeverityWarning }

// IsGlobal returns true if the error is a global error.
func (e Error) IsGlobal() bool { return e.global }

// Error implements the Error interface.
func (e Error) Error() string {
	return e.err
}

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
		return nil, ErrMessageEmpty
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
	if _, found := r.GetClient(u.ID); found {
		return ErrExistingSession
	}

	if r.IsAtCapacity() {
		return ErrRoomFull
	}

	r.muClients.Lock()
	r.clients[u.ID.String()] = &Client{
		user: u,
		conn: ws,
	}
	r.muClients.Unlock()

	return nil
}

// IsAtCapacity returns true if the room is at capacity.
func (r *Room) IsAtCapacity() bool {
	return r.NumUsers() >= uint64(maxClients)
}

// GetClient gets a client.
func (r *Room) GetClient(id xid.ID) (*Client, bool) {
	r.muClients.RLock()
	defer r.muClients.RUnlock()
	client, found := r.clients[id.String()]
	return client, found
}

// RemoveClient removes a client.
func (r *Room) RemoveClient(id xid.ID) bool {
	if _, found := r.GetClient(id); !found {
		return false
	}

	r.muClients.Lock()
	delete(r.clients, id.String())
	r.muClients.Unlock()

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
