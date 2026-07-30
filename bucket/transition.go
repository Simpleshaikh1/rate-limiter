package bucket

import "time"

func Transition(cfg Config, state State, now time.Time) Decision {
	next := state
	interval := cfg.RefillInterval()
	capacity := cfg.Capacity()
	tokens := state.Tokens

	elapsed := normalizeElapsed(now, state.LastRefill)

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

// helpers
func normalizeElapsed(
	now time.Time,
	last time.Time,
) time.Duration {

	elapsed := now.Sub(last)

	if elapsed < 0 {
		return 0
	}

	return elapsed
}
