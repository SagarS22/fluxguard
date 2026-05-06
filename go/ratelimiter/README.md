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

## Options

- `WithAlgorithm(algorithm)` selects a custom algorithm.
- `WithKeyPrefix(prefix)` overrides the Redis key prefix.
- `WithScriptPath(path)` loads a Lua script from a custom path.
- `WithScriptSource(source)` supplies Lua source directly, useful for tests.

## Registry / factory

```go
ratelimiter.Registry.Register("custom", func() ratelimiter.Algorithm { return CustomAlgorithm{} })
limiter, err := ratelimiter.CreateRateLimiter(redisClient, "custom")
```
