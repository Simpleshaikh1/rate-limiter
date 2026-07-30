package limiter

import (
	"sync"
	"time"

	"github.com/Simpleshaikh1/rate-limiter/bucket"
)

type Limiter struct {
	mu sync.Mutex

	cfg bucket.Config

	state bucket.State
}

func New(cfg bucket.Config) *Limiter {
	return &Limiter{
		cfg: cfg,
		state: bucket.State{
			Tokens:     cfg.Capacity(),
			LastRefill: time.Now(),
		},
	}
}

func (l *Limiter) Allow() bucket.Decision {
	l.mu.Lock()
	defer l.mu.Unlock()

	decision := bucket.Transition(
		l.cfg,
		l.state,
		time.Now(),
	)

	l.state = decision.State

	return decision
}
