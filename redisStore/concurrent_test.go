package redisstore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStore_ConcurrentSameBucket(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:concurrent"

	defer store.client.Del(ctx, key)

	const (
		capacity = 1000
		requests = 10_000
	)

	now := time.UnixMilli(1_000_000)

	var wg sync.WaitGroup

	var allowed atomic.Int64
	var rejected atomic.Int64

	wg.Add(requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			decision, err := store.Allow(
				ctx,
				key,
				capacity,
				time.Hour,
				now,
			)

			if err != nil {
				t.Errorf("Allow() error: %v", err)
				return
			}

			if decision.Allowed {
				allowed.Add(1)
			} else {
				rejected.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != capacity {
		t.Fatalf(
			"expected %d allowed, got %d",
			capacity,
			got,
		)
	}

	if got := rejected.Load(); got != requests-capacity {
		t.Fatalf(
			"expected %d rejected, got %d",
			requests-capacity,
			got,
		)
	}
}
