package bucket

import "time"

func Transition(cfg Config, state State, now time.Time) Decision {
	next := state

	elapsed := now.Sub(state.LastRefill)

	interval := elapsed / cfg.RefillInterval()

	remainder := elapsed % cfg.RefillInterval()

	_ = interval
	_ = remainder

	//if next.Tokens > 0 {
	//	next.Tokens--
	//
	//	return Decision{
	//		State:      next,
	//		Allowed:    true,
	//		RetryAfter: 0,
	//	}
	//}
	return Decision{
		State: next,
	}
}
