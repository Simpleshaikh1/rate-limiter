package redisstore

import (
	"context"
	"errors"
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

	ttl := 5 * time.Second

	got, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
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

	ttl := 5 * time.Second

	for i := 0; i < 10; i++ {
		decision, err := store.Allow(
			ctx,
			key,
			10,
			time.Second,
			ttl,
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
		ttl,
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

	ttl := 5 * time.Second

	// Consume one.
	first, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
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
		ttl,
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

	ttl := 5 * time.Second

	first, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
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
		ttl,
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

func TestStore_Allow_SetsTTL(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:ttl"

	defer store.client.Del(ctx, key)

	now := time.UnixMilli(1_000_000)

	ttl := 5 * time.Second

	_, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
		now,
	)

	if err != nil {
		t.Fatal(err)
	}

	actualTTL, err := store.client.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}

	if actualTTL <= 0 {
		t.Fatal("expected key to have a TTL")
	}

	if actualTTL > ttl {
		t.Fatalf(
			"TTL %v is greater than configured TTL %v",
			actualTTL,
			ttl,
		)
	}
}

func TestStore_Allow_RefreshesTTL(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	key := "test:rate-limit:ttl-refresh"

	defer store.client.Del(ctx, key)

	now := time.UnixMilli(1_000_000)

	ttl := 5 * time.Second

	_, err := store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
		now,
	)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	before, err := store.client.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Allow(
		ctx,
		key,
		10,
		time.Second,
		ttl,
		now.Add(100*time.Millisecond),
	)

	if err != nil {
		t.Fatal(err)
	}

	after, err := store.client.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}

	if after <= before {
		t.Fatalf(
			"expected TTL to refresh: before=%v after=%v",
			before,
			after,
		)
	}
}

func TestStore_Allow_ContextCancelled(t *testing.T) {
	store := newTestStore(t)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	_, err := store.Allow(
		ctx,
		"test:rate-limit:cancelled",
		10,
		time.Second,
		5*time.Second,
		time.Now(),
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}
