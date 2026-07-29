package bucket

import "time"

func Transition(cfg Config, state State, now time.Time) Decision {
	next := state
	interval := cfg.RefillInterval()
	capacity := cfg.Capacity()
	tokens := state.Tokens

	elapsed := now.Sub(state.LastRefill)

	if elapsed < 0 {
		elapsed = 0
	}

	intervals := elapsed / interval
	remainder := elapsed % interval
	retryAfter := interval - remainder

	tokensToAdd := int(intervals)
	tokens += tokensToAdd

	if tokens > capacity {
		tokens = capacity
	}

	next.LastRefill = state.LastRefill.Add(intervals * interval)

	if tokens > 0 {
		tokens--

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
