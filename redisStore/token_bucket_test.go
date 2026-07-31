package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis unavailable: %v", err)
	}

	return New(client)
}

func TestStore_Allow_FirstRequest(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:first-request"

	defer store.client.Del(ctx, key)

	now := time.UnixMilli(1_000_000)

	got, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Allowed {
		t.Fatal("first request should be allowed")
	}

	if got.Tokens != 9 {
		t.Fatalf(
			"expected 9 tokens, got %d",
			got.Tokens,
		)
	}

	if !got.RetryAfter.IsZero() {
		t.Fatalf(
			"expected zero retry duration, got %v",
			got.RetryAfter,
		)
	}
}

func TestStore_Allow_RejectsEmptyBucket(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:reject"

	defer store.client.Del(ctx, key)

	now := time.UnixMilli(1_000_000)

	for i := 0; i < 10; i++ {
		decision, err := store.Allow(
			ctx,
			key,
			10,
			time.Second,
			now,
		)

		if err != nil {
			t.Fatal(err)
		}

		if !decision.Allowed {
			t.Fatalf(
				"request %d should be allowed",
				i+1,
			)
		}
	}

	decision, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		now,
	)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {
		t.Fatal("11th request should be rejected")
	}

	if decision.Tokens != 0 {
		t.Fatalf(
			"expected 0 tokens, got %d",
			decision.Tokens,
		)
	}
}

func TestStore_Allow_Refills(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:refill"

	defer store.client.Del(ctx, key)

	base := time.UnixMilli(1_000_000)

	// Consume one.
	first, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		base,
	)
	if err != nil {
		t.Fatal(err)
	}

	if first.Tokens != 9 {
		t.Fatalf("expected 9, got %d", first.Tokens)
	}

	// One second later:
	second, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		base.Add(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	/*
		9 + 1 refill - 1 consume = 9
	*/
	if second.Tokens != 9 {
		t.Fatalf(
			"expected 9 tokens after refill, got %d",
			second.Tokens,
		)
	}

	if !second.Allowed {
		t.Fatal("request should be allowed")
	}
}

func TestStore_Allow_PreservesPartialInterval(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:partial"

	defer store.client.Del(ctx, key)

	base := time.UnixMilli(1_000_000)

	first, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		base,
	)
	if err != nil {
		t.Fatal(err)
	}

	if first.Tokens != 9 {
		t.Fatalf("expected 9, got %d", first.Tokens)
	}

	second, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		base.Add(1500*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}

	/*
		9 + 1 refill - 1 consume = 9
	*/

	if second.Tokens != 9 {
		t.Fatalf(
			"expected 9 tokens, got %d",
			second.Tokens,
		)
	}

	expectedRefill := base.Add(time.Second)

	if !second.LastRefill.Equal(expectedRefill) {
		t.Fatalf(
			"expected last refill %v, got %v",
			expectedRefill,
			second.LastRefill,
		)
	}
}
