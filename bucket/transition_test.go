package bucket

import (
	"testing"
	"time"
)

func TestTransition_ConsumeToken(t *testing.T) {
	base := time.Unix(1_000_000, 0)

	cfg := mustConfig(10, time.Second)

	state := State{
		Tokens:     5,
		LastRefill: base,
	}

	want := Decision{
		State: State{
			Tokens:     4,
			LastRefill: base,
		},
		Allowed:    true,
		RetryAfter: 0,
	}

	got := Transition(cfg, state, base)

	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// Helper to create configs in tests, panics on invalid input
func mustConfig(capacity int, interval time.Duration) Config {
	cfg, err := NewConfig(capacity, interval)
	if err != nil {
		panic(err)
	}
	return cfg
}
