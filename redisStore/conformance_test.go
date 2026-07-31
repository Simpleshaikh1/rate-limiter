package redisstore

import (
	"context"
	"github.com/Simpleshaikh1/rate-limiter/bucket"
	"testing"
	"time"
)

func TestRedisMatchesTransition(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	const (
		capacity = 10
		interval = time.Second
	)

	cfg, err := bucket.NewConfig(capacity, interval)
	if err != nil {
		t.Fatal(err)
	}

	base := time.UnixMilli(1_000_000)

	tests := []struct {
		name  string
		state bucket.State
		now   time.Time
	}{
		{
			name: "available tokens",
			state: bucket.State{
				Tokens:     5,
				LastRefill: base,
			},
			now: base,
		},
		{
			name: "one interval",
			state: bucket.State{
				Tokens:     5,
				LastRefill: base,
			},
			now: base.Add(time.Second),
		},
		{
			name: "multiple intervals",
			state: bucket.State{
				Tokens:     5,
				LastRefill: base,
			},
			now: base.Add(3 * time.Second),
		},
		{
			name: "partial interval",
			state: bucket.State{
				Tokens:     5,
				LastRefill: base,
			},
			now: base.Add(1500 * time.Millisecond),
		},
		{
			name: "empty bucket",
			state: bucket.State{
				Tokens:     0,
				LastRefill: base,
			},
			now: base.Add(300 * time.Millisecond),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := "test:conformance:" + tc.name

			defer store.client.Del(ctx, key)

			/*
				Seed Redis with exactly the same
				state used by the Go implementation.
			*/
			err := store.client.HSet(
				ctx,
				key,
				"tokens",
				tc.state.Tokens,
				"last_refill_ms",
				tc.state.LastRefill.UnixMilli(),
			).Err()

			if err != nil {
				t.Fatal(err)
			}

			// Run Go implementation.
			expected := bucket.Transition(
				cfg,
				tc.state,
				tc.now,
			)

			// Run Redis implementation.
			actual, err := store.Allow(
				ctx,
				key,
				capacity,
				interval,
				tc.now,
			)

			if err != nil {
				t.Fatal(err)
			}

			if actual.Allowed != expected.Allowed {
				t.Fatalf(
					"Allowed mismatch: got %v, want %v",
					actual.Allowed,
					expected.Allowed,
				)
			}

			if actual.Tokens != expected.State.Tokens {
				t.Fatalf(
					"Tokens mismatch: got %d, want %d",
					actual.Tokens,
					expected.State.Tokens,
				)
			}

			if !actual.LastRefill.Equal(
				expected.State.LastRefill,
			) {
				t.Fatalf(
					"LastRefill mismatch: got %v, want %v",
					actual.LastRefill,
					expected.State.LastRefill,
				)
			}

			if actual.RetryAfter != expected.RetryAfter {
				t.Fatalf(
					"RetryAfter mismatch: got %v, want %v",
					actual.RetryAfter,
					expected.RetryAfter,
				)
			}
		})
	}
}
