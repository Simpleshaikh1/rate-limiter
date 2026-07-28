package bucket

import "time"

type Decision struct {
	State State

	Allowed bool

	RetryAfter time.Duration
}
