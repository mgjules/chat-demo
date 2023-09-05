package main

import (
	"context"
	"fmt"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/rs/xid"
)

type userContextKey string

const userCtxKey userContextKey = "user"

// user holds information about a user.
type user struct {
	ID   xid.ID
	Name string
}

// newUser creates a new user.
func newUser() *user {
	id := xid.New()
	// Prevents faker from tracking duplicates since it does that in a non-treadsafe manner.
	// Instead we seed the Name with sections of the ID.
	nonunique := options.WithGenerateUniqueValues(false)
	return &user{
		ID: id,
		Name: fmt.Sprintf("%s %s (%s%s)",
			faker.FirstName(nonunique), faker.LastName(nonunique), id.String()[4:8], id.String()[15:],
		),
	}
}

func addUserToContext(ctx context.Context, user *user) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func userFromContext(ctx context.Context) *user {
	u, ok := ctx.Value(userCtxKey).(*user)
	if !ok {
		return nil
	}

	return u
}
