package redisstore

import "time"

type Decision struct {
	Allowed    bool
	Tokens     int
	LastRefill time.Time
	RetryAfter time.Duration
}
