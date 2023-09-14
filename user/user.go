package user

import (
	"context"
	"fmt"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/rs/xid"
)

type userContextKey string

const userCtxKey userContextKey = "user"

// User holds information about a user.
type User struct {
	ID   xid.ID
	Name string
}

// New creates a new User.
func New() *User {
	id := xid.New()
	// Prevents faker from tracking duplicates since it does that in a non-threadsafe manner.
	// Instead we seed the Name with sections of the ID.
	nonunique := options.WithGenerateUniqueValues(false)
	return &User{
		ID: id,
		Name: fmt.Sprintf("%s %s (%s%s)",
			faker.FirstName(nonunique), faker.LastName(nonunique), id.String()[4:8], id.String()[15:],
		),
	}
}

// AddToContext adds a user to the context.
func AddToContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// FromContext retrieves a user from the context.
func FromContext(ctx context.Context) *User {
	u, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil
	}

	return u
}
