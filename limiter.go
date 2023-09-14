package main

import (
	"sync"
	"time"

	"github.com/mgjules/chat-demo/user"
	"golang.org/x/time/rate"
)

type limiters struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
}

func newLimiters() *limiters {
	return &limiters{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (l *limiters) add(u *user.User, d time.Duration, b int) *rate.Limiter {
	l.mu.RLock()
	limiter, found := l.limiters[u.ID.String()]
	l.mu.RUnlock()
	if found {
		return limiter
	}

	limiter = rate.NewLimiter(rate.Every(d), b)
	l.mu.Lock()
	l.limiters[u.ID.String()] = limiter
	l.mu.Unlock()

	return limiter
}

func (l *limiters) remove(u *user.User) {
	l.mu.Lock()
	delete(l.limiters, u.ID.String())
	l.mu.Unlock()
}
