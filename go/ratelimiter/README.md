# Fluxguard Go Rate Limiter SDK

Extensible Redis/Lua-backed rate limiter SDK for Go.

## Default token bucket

```go
limiter, err := ratelimiter.New(redisClient)
if err != nil { panic(err) }

decision, err := limiter.Check(ctx, "user:123", ratelimiter.TokenBucketPolicy{
    Capacity: 100,
    RefillRate: 10,
    TTLSec: 60,
}, 1)
```

`redisClient` must satisfy:

```go
type RedisClient interface {
    ScriptLoad(ctx context.Context, script string) (string, error)
    EvalSHA(ctx context.Context, sha string, keys []string, args ...any) (any, error)
}
```

Use a small adapter for your Redis library if needed.

## Sliding window

```go
limiter, err := ratelimiter.CreateRateLimiter(redisClient, "sliding_window")
if err != nil { panic(err) }

decision, err := limiter.Check(ctx, "user:123", ratelimiter.SlidingWindowPolicy{
    Capacity: 100,
    WindowSec: 60,
    TTLSec: 120,
}, 1)
```

The sliding-window algorithm uses two Redis keys: the request log key and a sequence key. The SDK wraps the encoded user key in a Redis Cluster hash tag so both keys share a hash slot, for example:

```text
ratelimit:sliding_window:key:{dXNlcjoxMjM}
ratelimit:sliding_window:key:{dXNlcjoxMjM}:seq
```

## Options

- `WithAlgorithm(algorithm)` selects a custom algorithm.
- `WithKeyPrefix(prefix)` overrides the logical Redis key prefix.
- `WithScriptPath(path)` loads a Lua script from a custom path.
- `WithScriptSource(source)` supplies Lua source directly, useful for tests.

## Registry / factory

```go
ratelimiter.Registry.Register("custom", func() ratelimiter.Algorithm { return CustomAlgorithm{} })
limiter, err := ratelimiter.CreateRateLimiter(redisClient, "custom")
```
