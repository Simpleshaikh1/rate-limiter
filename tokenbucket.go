package rate_limiter_token_bucket_leaky_bucket

import (
	"fmt"
	"time"
)

type TokenBucket struct {
	//Configuration
	capacity       int
	refillInterval time.Duration

	//Mutable state
	tokens     int
	lastRefill time.Time

	//Dependencies
	clock Clock
}

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
