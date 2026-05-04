# fluxguard-ratelimiter (Python)

Python SDK for Redis Lua-based rate limiting with extensible algorithms.

## Install

```bash
pip install -e .
```

## Usage (default token bucket)

```python
import redis
from ratelimiter import RateLimiter, Policy

r = redis.Redis.from_url("redis://localhost:6379/0")
rl = RateLimiter(r)

decision = rl.check(
    key="user:123",
    policy=Policy(capacity=10, refill_rate=5.0, ttl_sec=60),
    requested=1,
)

print(decision.allowed, decision.remaining, decision.retry_after_ms, decision.reset_at_ms)
```

## Usage (factory)

```python
from ratelimiter import create_rate_limiter

rl = create_rate_limiter(r, algorithm="token_bucket")
```

## Extending with custom algorithm

Implement the `RateLimitAlgorithm` contract and register it:

```python
from ratelimiter import registry

registry.register("my_algo", MyAlgorithm)
rl = create_rate_limiter(r, algorithm="my_algo")
```

## Decision fields

- `allowed`: request can consume tokens.
- `remaining`: integer tokens left after the check.
- `retry_after_ms`: wait time when denied.
- `reset_at_ms`: epoch milliseconds when bucket reaches full state.
