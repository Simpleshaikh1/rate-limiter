package bucket

import "time"

func Transition(cfg Config, state State, now time.Time) Decision {
	next := state

	elapsed := now.Sub(state.LastRefill)

	intervals := elapsed / cfg.RefillInterval()

	remainder := elapsed % cfg.RefillInterval()

	retryAfter := cfg.RefillInterval() - remainder

	tokensToAdd := int(intervals)

	next.Tokens += tokensToAdd

	if next.Tokens > cfg.Capacity() {
		next.Tokens = cfg.Capacity()
	}

	next.LastRefill = state.LastRefill.Add(intervals * cfg.RefillInterval())

	if next.Tokens > 0 {
		next.Tokens--

		return Decision{
			State:      next,
			Allowed:    true,
			RetryAfter: 0,
		}
	}
	return Decision{
		State: next,

		Allowed: false,

		RetryAfter: retryAfter,
	}
}
