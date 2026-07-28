package bucket

import (
	"fmt"
	"time"
)

type Config struct {
	// private fields
}

func NewConfig(
	capacity int,
	refillInterval time.Duration,
) (Config, error)

func (c Config) Capacity() int
func (c Config) RefillInterval() time.Duration

type State struct {
	Tokens     int
	LastRefill time.Time
}

type Decision struct {
	State      State
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

func Transition(
	cfg Config,
	state State,
	now time.Time,
) Decision

func NewTokenBucket(capacity, refillRate int) (*TokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than zero")
	}

	if refillRate <= 0 {
		return nil, fmt.Errorf("refill rate must be geater than zero")
	}

	//clock := SystemClock{}
	return newTokenBucket(
		capacity,
		refillRate,
		SystemClock{},
	), nil
}

func newTokenBucket(
	capacity,
	refillRate int,
	clock Clock,
) (*TokenBucket, error) {

	if capacity <= 0 {
		return nil, fmt.Errorf(
			"capacity must be greater than zero",
		)
	}

	if refillRate <= 0 {
		return nil, fmt.Errorf(
			"refill rate must be greater than zero",
		)
	}

	if clock == nil {
		return nil, fmt.Errorf(
			"clock cannot be nil",
		)
	}

	return &TokenBucket{
		capacity: capacity,

		refillRate: refillRate,

		tokens: capacity,

		lastRefill: clock.Now(),

		clock: clock,
	}, nil
}

func (b *TokenBucket) Allow() bool {
	now := b.clock.Now()

	elapsed := now.Sub(b.lastRefill)

	intervals := elapsed / b.refillInterval

	if intervals > 0 {
		b.tokens += int(intervals)

		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}

		b.lastRefill = b.lastRefill.Add(
			intervals * b.refillInterval,
		)
	}

	if b.tokens == 0 {
		return false
	}

	b.tokens--

	return true
}

// Pure function
func Transition(cfg BucketConfig, state BucketState, now time.Time) (BucketState, bool) {
	next := state

	elapsed := now.Sub(state.LastRefill)

	intervals := elapsed / cfg.RefillInterval

	if intervals > 0 {
		next.Tokens += int(intervals)

		if next.Tokens > cfg.Capacity {
			next.Tokens = cfg.Capacity
		}

		next.LastRefill = state.LastRefill.Add(intervals * cfg.RefillInterval)

		if next.Tokens == 0 {
			return next, false
		}

		next.Tokens--

		return next, true
	}
}
