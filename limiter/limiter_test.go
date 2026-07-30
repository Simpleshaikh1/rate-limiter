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

	if len(l.shards) != workers {
		t.Fatalf(
			"expected %d buckets, got %d",
			workers,
			len(l.shards),
		)
	}
}

func BenchmarkLimiter(b *testing.B) {
	cfg, _ := bucket.NewConfig(1000, time.Second)

	l := New(cfg)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Allow("client")
		}
	})
}

func TestLimiter_ClientIsolation(t *testing.T) {
	cfg, err := bucket.NewConfig(1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	l := New(cfg)

	if !l.Allow("alice").Allowed {
		t.Fatal("first Alice request should be allowed")
	}

	if l.Allow("alice").Allowed {
		t.Fatal("second Alice request should be rejected")
	}

	if !l.Allow("bob").Allowed {
		t.Fatal("first Bob request should be allowed")
	}
}

func TestLimiter_ConcurrentSameClient(t *testing.T) {
	cfg, err := bucket.NewConfig(10_000, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	l := New(cfg)

	const requests = 10_000

	var wg sync.WaitGroup
	var allowed atomic.Int64

	wg.Add(requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			if l.Allow("client-1").Allowed {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != requests {
		t.Fatalf(
			"expected %d allowed requests, got %d",
			requests,
			got,
		)
	}
}

func TestLimiter_ConcurrentRejectsAfterCapacity(t *testing.T) {
	cfg, err := bucket.NewConfig(100, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	l := New(cfg)

	const requests = 10_000

	var wg sync.WaitGroup
	var allowed atomic.Int64
	var rejected atomic.Int64

	wg.Add(requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			decision := l.Allow("client-1")

			if decision.Allowed {
				allowed.Add(1)
			} else {
				rejected.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Fatalf("expected 100 allowed, got %d", got)
	}

	if got := rejected.Load(); got != 9_900 {
		t.Fatalf("expected 9900 rejected, got %d", got)
	}
}

func TestLimiter_ConcurrentMultipleClients(t *testing.T) {
	cfg, err := bucket.NewConfig(100, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	l := New(cfg)

	const clients = 100
	const requestsPerClient = 100

	var wg sync.WaitGroup

	wg.Add(clients * requestsPerClient)

	for i := 0; i < clients; i++ {
		clientID := fmt.Sprintf("client-%d", i)

		for j := 0; j < requestsPerClient; j++ {
			go func(id string) {
				defer wg.Done()

				l.Allow(id)
			}(clientID)
		}
	}

	wg.Wait()
}
