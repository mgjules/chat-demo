package main

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiters struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func newLimiters() *limiters {
	return &limiters{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (l *limiters) add(u *user, d time.Duration, b int) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, found := l.limiters[u.ID.String()]; found {
		return limiter
	}
	limiter := rate.NewLimiter(rate.Every(d), b)
	l.limiters[u.ID.String()] = limiter

	return limiter
}

func (l *limiters) remove(u *user) {
	l.mu.Lock()
	delete(l.limiters, u.ID.String())
	l.mu.Unlock()
}
