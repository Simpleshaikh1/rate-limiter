package limiter

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/Simpleshaikh1/rate-limiter/bucket"
)

type Limiter struct {
	cfg bucket.Config

	shards []shard
}

type shard struct {
	mu sync.Mutex

	buckets map[string]bucket.State
}

const shardCount = 64

func New(cfg bucket.Config) *Limiter {
	l := &Limiter{
		cfg:    cfg,
		shards: make([]shard, shardCount),
	}

	for i := range l.shards {
		l.shards[i].buckets = make(map[string]bucket.State)
	}

	return l
}

func (l *Limiter) Allow(clientID string) bucket.Decision {
	s := l.shard(clientID)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	state := l.loadBucket(clientID, now)

	decision := bucket.Transition(
		l.cfg,
		state,
		time.Now(),
	)

	s.buckets[clientID] = decision.State

	return decision
}

func (l *Limiter) loadBucket(clientID string, now time.Time) bucket.State {
	s := l.shard(clientID)
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

func (l *Limiter) shard(clientID string) *shard {
	h := fnv.New32a()

	_, _ = h.Write([]byte(clientID))

	index := h.Sum32() % uint32(len(l.shards))

	return &l.shards[index]
}
