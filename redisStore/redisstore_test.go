package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestStore_SetGet(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	store := New(client)

	err := store.Set(ctx, "test:key", "hello")
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.Get(ctx, "test:key")
	if err != nil {
		t.Fatal(err)
	}

	if got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}
