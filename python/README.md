# fluxguard-ratelimiter (Python)

Python SDK for the Redis Lua token bucket limiter.

## Install

```bash
pip install -e .
```

## Usage

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

## Decision fields

- `allowed`: request can consume tokens.
- `remaining`: integer tokens left after the check.
- `retry_after_ms`: wait time when denied.
- `reset_at_ms`: epoch milliseconds when bucket reaches full state.
