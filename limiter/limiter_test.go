package limiter

import (
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

			d := l.Allow()

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
