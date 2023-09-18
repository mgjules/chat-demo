package main

import (
	"sync"
	"time"

	mlimiters "github.com/mennanov/limiters"
	"github.com/mgjules/chat-demo/user"
)

type limiters struct {
	mu       sync.RWMutex
	limiters map[string]*mlimiters.TokenBucket
}

func newLimiters() *limiters {
	return &limiters{
		limiters: make(map[string]*mlimiters.TokenBucket),
	}
}

func (l *limiters) add(u *user.User, d time.Duration, b int64) *mlimiters.TokenBucket {
	l.mu.RLock()
	limiter, found := l.limiters[u.ID.String()]
	l.mu.RUnlock()
	if found {
		return limiter
	}

	limiter = mlimiters.NewTokenBucket(b, d, mlimiters.NewLockNoop(), mlimiters.NewTokenBucketInMemory(), mlimiters.NewSystemClock(), mlimiters.NewStdLogger())
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
