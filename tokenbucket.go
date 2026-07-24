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
}

func NewTokenBucket(capacity, refillRate int) (*TokenBucket, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than zero")
	}

	if refillRate <= 0 {
		return nil, fmt.Errorf("refill rate must be geater than zero")
	}

	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}, nil
}
