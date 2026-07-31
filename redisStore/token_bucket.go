package redisstore

import "github.com/redis/go-redis/v9"

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_interval_ms = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])

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

return {
    allowed,
    tokens,
    last_refill_ms,
    retry_after_ms
}
`)
