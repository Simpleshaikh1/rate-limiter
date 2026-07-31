package redisstore

import (
	"context"
	"github.com/redis/go-redis/v9"
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
