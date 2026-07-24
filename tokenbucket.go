package rate_limiter_token_bucket_leaky_bucket

import (
	"fmt"
	"time"
)

type TokenBucket struct {
	capacity   int
	refillRate int

	tokens     int
	lastRefill time.Time

	clock Clock
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

func NewTokenBucket(capacity, refillRate int) (*TokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than zero")
	}

	if refillRate <= 0 {
		return nil, fmt.Errorf("refill rate must be geater than zero")
	}

	clock := SystemClock{}
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: clock.Now(),
		clock:      clock,
	}, nil
}
