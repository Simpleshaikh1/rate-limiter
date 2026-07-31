package redisstore

import (
	"context"
	"fmt"
	"github.com/Simpleshaikh1/rate-limiter/bucket"
	"github.com/redis/go-redis/v9"
	"time"
)

type Store struct {
	client *redis.Client
}

func New(client *redis.Client) *Store {
	return &Store{
		client: client,
	}
}

func (s *Store) Set(
	ctx context.Context,
	key string,
	value string,
) error {
	return s.client.Set(ctx, key, value, 0).Err()
}

func (s *Store) Get(
	ctx context.Context,
	key string,
) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *Store) Allow(
	ctx context.Context,
	key string,
	capacity int,
	refillInterval time.Duration,
	now time.Time,
) (bucket.Decision, error) {
	result, err := tokenBucketScript.Run(
		ctx,
		s.client,
		[]string{key},
		capacity,
		refillInterval.Milliseconds(),
		now.UnixMilli(),
	).Result()

	if err != nil {
		return bucket.Decision{}, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		return bucket.Decision{}, fmt.Errorf("unexpected Redis response: %T", result)
	}

	// Parse values...
}

//func test() {
//	ctx := context.Background()
//
//	client := redis.NewClient(&redis.Options{
//		Addr: "localhost:6379",
//	})
//
//	err := client.Ping(ctx).Err()
//	if err != nil {
//		panic(err)
//	}
//}
