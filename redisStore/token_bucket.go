package redisstore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_interval_ms = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])
local ttl_ms = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_refill_ms")

local tokens = tonumber(data[1])
local last_refill_ms = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last_refill_ms = now_ms
end

local elapsed_ms = now_ms - last_refill_ms

if elapsed_ms >= refill_interval_ms then
    local intervals = math.floor(elapsed_ms / refill_interval_ms)

    tokens = math.min(
        capacity,
        tokens + intervals
    )

    last_refill_ms =
        last_refill_ms +
        intervals * refill_interval_ms
end

local allowed = 0
local retry_after_ms = 0

if tokens > 0 then
    tokens = tokens - 1
    allowed = 1
else
    local remaining_ms =
        refill_interval_ms - (now_ms - last_refill_ms)

    retry_after_ms = math.max(0, remaining_ms)
end

redis.call(
    "HSET",
    key,
    "tokens",
    tokens,
    "last_refill_ms",
    last_refill_ms
)
redis.call(
    "PEXPIRE",
    key,
    ttl_ms
)

return {
    allowed,
    tokens,
    last_refill_ms,
    retry_after_ms
}
`)

func (s *Store) Allow(
	ctx context.Context,
	key string,
	capacity int,
	refillInterval time.Duration,
	ttl time.Duration,
	now time.Time,
) (Decision, error) {
	result, err := tokenBucketScript.Run(
		ctx,
		s.client,
		[]string{key},
		capacity,
		refillInterval.Milliseconds(),
		now.UnixMilli(),
	).Result()

	if ttl <= 0 {
		return Decision{}, fmt.Errorf(
			"ttl must be greater than zero",
		)
	}

	if capacity <= 0 {
		return Decision{}, fmt.Errorf(
			"capacity must be greater than zero",
		)
	}

	if err != nil {
		return Decision{}, err
	}

	values, ok := result.([]interface{})
	if !ok {
		return Decision{}, fmt.Errorf(
			"unexpected Redis response type: %T",
			result,
		)
	}

	if len(values) != 4 {
		return Decision{}, fmt.Errorf(
			"unexpected Redis response length: %d",
			len(values),
		)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return Decision{}, fmt.Errorf(
			"parse allowed: %w",
			err,
		)
	}

	tokens, err := redisInt(values[1])
	if err != nil {
		return Decision{}, fmt.Errorf(
			"parse tokens: %w",
			err,
		)
	}

	lastRefillMs, err := redisInt64(values[2])
	if err != nil {
		return Decision{}, fmt.Errorf(
			"parse last_refill_ms: %w",
			err,
		)
	}

	retryAfterMs, err := redisInt64(values[3])
	if err != nil {
		return Decision{}, fmt.Errorf(
			"parse retry_after_ms: %w",
			err,
		)
	}

	return Decision{
		Allowed:    allowed == 1,
		Tokens:     tokens,
		LastRefill: time.UnixMilli(lastRefillMs),
		RetryAfter: time.Duration(retryAfterMs) * time.Millisecond,
	}, nil
}

func redisInt(v interface{}) (int, error) {
	n, err := redisInt64(v)
	if err != nil {
		return 0, err
	}

	return int(n), nil
}

func redisInt64(v interface{}) (int64, error) {
	switch value := v.(type) {
	case int64:
		return value, nil

	case int:
		return int64(value), nil

	case string:
		var n int64

		_, err := fmt.Sscanf(value, "%d", &n)
		if err != nil {
			return 0, fmt.Errorf(
				"invalid integer %q",
				value,
			)
		}

		return n, nil

	default:
		return 0, fmt.Errorf(
			"unexpected integer type %T",
			v,
		)
	}
}
