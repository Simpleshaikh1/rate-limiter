package bucket

import (
	"errors"
	"time"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than zero")
var ErrInvalidRefillInterval = errors.New("refill interval must be greater than zero")

type Config struct {
	capacity       int
	refillInterval time.Duration
}

func NewConfig(capacity int, refillInterval time.Duration) (Config, error) {
	if capacity <= 0 {
		return Config{}, ErrInvalidCapacity
	}

	if refillInterval <= 0 {
		return Config{}, ErrInvalidRefillInterval
	}

	return Config{
		capacity:       capacity,
		refillInterval: refillInterval,
	}, nil
}

func (c Config) Capacity() int {
	return c.capacity
}

func (c Config) RefillInterval() time.Duration {
	return c.refillInterval
}
