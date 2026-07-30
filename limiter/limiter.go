package limiter

import (
	"sync"
	"time"

	"github.com/Simpleshaikh1/rate-limiter/bucket"
)

type Limiter struct {
	mu sync.Mutex

	cfg bucket.Config

	buckets map[string]bucket.State
}

func New(cfg bucket.Config) *Limiter {
	return &Limiter{
		cfg:     cfg,
		buckets: make(map[string]bucket.State),
	}
}

func (l *Limiter) Allow(clientID string) bucket.Decision {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	state := l.loadBucket(clientID, now)

	decision := bucket.Transition(
		l.cfg,
		state,
		time.Now(),
	)

	l.buckets[clientID] = decision.State

	return decision
}

func (l *Limiter) loadBucket(clientID string, now time.Time) bucket.State {
	state, ok := l.buckets[clientID]
	if ok {
		return state
	}

	return bucket.State{
		Tokens:     l.cfg.Capacity(),
		LastRefill: now,
	}
}
