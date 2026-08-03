package limiter

import (
	"errors"
	"hash/fnv"
	"sync"
	"time"

	"github.com/Simpleshaikh1/rate-limiter/bucket"
)

var (
	ErrInvalidCapacity       = errors.New("capacity must be greater than zero")
	ErrInvalidRefillInterval = errors.New("refill interval must be greater than zero")
	ErrInvalidTTL            = errors.New("ttl must be greater than zero")
)

const shardCount = 64

type Config struct {
	Capacity       int
	RefillInterval time.Duration
	TTL            time.Duration
}

func (c Config) Validate() error {
	if c.Capacity <= 0 {
		return ErrInvalidCapacity
	}

	if c.RefillInterval <= 0 {
		return ErrInvalidRefillInterval
	}

	if c.TTL <= 0 {
		return ErrInvalidTTL
	}

	return nil
}

type shard struct {
	mu sync.Mutex

	buckets map[string]bucket.State
}

type Limiter struct {
	cfg bucket.Config

	shards []shard
}

func New(cfg bucket.Config) *Limiter {
	shards := make([]shard, shardCount)

	for i := range shards {
		shards[i].buckets = make(map[string]bucket.State)
	}
	l := &Limiter{
		cfg:    cfg,
		shards: shards,
	}

	return l
}

func (l *Limiter) getShard(clientID string) *shard {
	h := fnv.New32a()

	_, _ = h.Write([]byte(clientID))

	index := h.Sum32() % uint32(len(l.shards))

	return &l.shards[index]
}

func (l *Limiter) Allow(clientID string) bucket.Decision {
	s := l.getShard(clientID)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	state, exists := s.buckets[clientID]

	if !exists {
		state = bucket.State{
			Tokens:     l.cfg.Capacity(),
			LastRefill: now,
		}
	}

	decision := bucket.Transition(
		l.cfg,
		state,
		now,
	)

	s.buckets[clientID] = decision.State

	return decision
}

func (l *Limiter) loadBucket(clientID string, now time.Time) bucket.State {
	s := l.getShard(clientID)
	state, ok := s.buckets[clientID]
	if ok {
		return state
	}

	return bucket.State{
		Tokens:     l.cfg.Capacity(),
		LastRefill: now,
	}
}

//shard helper
