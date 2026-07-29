package bucket

import "time"

func Transition(cfg Config, state State, now time.Time) Decision {
	next := state

	if next.Tokens > 0 {
		next.Tokens--

		return Decision{
			State:      next,
			Allowed:    true,
			RetryAfter: 0,
		}
	}
	return Decision{
		State:      next,
		Allowed:    false,
		RetryAfter: cfg.RefillInterval(),
	}
}
