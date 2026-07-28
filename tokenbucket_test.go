package bucket

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tests := []struct {
		name       string
		capacity   int
		refillRate int
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			capacity:   10,
			refillRate: 5,
			wantErr:    false,
		},
		{
			name:       "zero capacity",
			capacity:   0,
			refillRate: 5,
			wantErr:    true,
		},
		{
			name:       "negative capacity",
			capacity:   -1,
			refillRate: 5,
			wantErr:    true,
		},
		{
			name:       "zero refill rate",
			capacity:   10,
			refillRate: 0,
			wantErr:    true,
		},
		{
			name:       "negative refill rate",
			capacity:   10,
			refillRate: -5,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bucket, err := NewTokenBucket(tc.capacity, tc.refillRate)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error but got nil")
				}
				if bucket != nil {
					t.Fatalf("expected nil bucket on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if bucket.capacity != tc.capacity {
				t.Fatalf("capacity = %d, want %d", bucket.capacity, tc.capacity)
			}

			if bucket.tokens != tc.capacity {
				t.Fatalf("tokens = %d, want %d", bucket.tokens, tc.capacity)
			}

			if bucket.refillRate != tc.refillRate {
				t.Fatalf("refillRate = %d, want %d", bucket.refillRate, tc.refillRate)
			}

			if bucket.lastRefill.IsZero() {
				t.Fatalf("lastRefill should be initialized")
			}
		})
	}
}

func TestAllow_ConsumesAvailableTokens(t *testing.T) {
	clock := NewFakeClock(time.Unix(0, 0))

	bucket, err := newTokenBucket(3, 1, clock)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !bucket.Allow() {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}

	if bucket.Allow() {
		t.Fatal("expected fourth request to be rejected")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	clock := NewFakeClock(time.Unix(0, 0))

	bucket, err := newTokenBucket(2, 1, clock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bucket.Allow() {
		t.Fatal("expected first request")
	}

	if !bucket.Allow() {
		t.Fatal("expected second request")
	}

	if bucket.Allow() {
		t.Fatal("bucket should be empty")
	}

	clock.Advance(time.Second)

	if !bucket.Allow() {
		t.Fatal("expected one refilled token")
	}

	if bucket.Allow() {
		t.Fatal("only one token should have been refilled")
	}
}

func TestAllow_DoesNotExceedCapacity(t *testing.T) {
	clock := NewFakeClock(time.Unix(0, 0))

	bucket, _ := newTokenBucket(5, 1, clock)

	for i := 0; i < 5; i++ {
		bucket.Allow()
	}

	clock.Advance(time.Hour)

	for i := 0; i < 5; i++ {
		if !bucket.Allow() {
			t.Fatal("expected token")
		}
	}

	if bucket.Allow() {
		t.Fatal("bucket exceeded capacity")
	}
}

func TestAllow_Concurrent(t *testing.T) {
	clock := NewFakeClock(time.Unix(0, 0))

	bucket, err := newTokenBucket(100, 1, clock)
	if err != nil {
		t.Fatal(err)
	}

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if bucket.Allow() {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	if allowed.Load() != 100 {
		t.Fatalf(
			"allowed=%d want=100",
			allowed.Load(),
		)
	}
}
