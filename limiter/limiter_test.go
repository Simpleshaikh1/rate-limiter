package limiter

import (
	"fmt"
	"github.com/Simpleshaikh1/rate-limiter/bucket"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiterConcurrent(t *testing.T) {
	cfg, _ := bucket.NewConfig(10000, time.Hour)

	l := New(cfg)

	const workers = 10000

	var wg sync.WaitGroup

	wg.Add(workers)

	var allowed atomic.Int64

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			d := l.Allow("")

			if d.Allowed {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	if allowed.Load() != workers {
		t.Fatalf(
			"expected %d allowed, got %d",
			workers,
			allowed.Load(),
		)
	}
}

func TestDifferentClientsHaveDifferentBuckets(t *testing.T) {
	cfg, _ := bucket.NewConfig(1, time.Hour)

	l := New(cfg)

	first := l.Allow("alice")
	if !first.Allowed {
		t.Fatal("alice should be allowed")
	}

	second := l.Allow("alice")
	if second.Allowed {
		t.Fatal("alice should be rate limited")
	}

	third := l.Allow("bob")
	if !third.Allowed {
		t.Fatal("bob should have an independent bucket")
	}
}

func TestConcurrentClients(t *testing.T) {
	cfg, _ := bucket.NewConfig(1000, time.Hour)

	l := New(cfg)

	const workers = 1000

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			clientID := fmt.Sprintf("client-%d", id)

			l.Allow(clientID)
		}(i)
	}

	wg.Wait()

	if len(l.buckets) != workers {
		t.Fatalf(
			"expected %d buckets, got %d",
			workers,
			len(l.buckets),
		)
	}
}
