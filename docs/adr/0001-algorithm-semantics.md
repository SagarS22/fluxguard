# ADR 0001: Algorithm Semantics

## Status
**Accepted** | Sprint 1

## Context
We need to define precise semantics for three rate limiting algorithms (Token Bucket, Sliding Window, Leaky Bucket) that will be implemented as Redis Lua scripts. The key challenge is ensuring atomic operations and consistent behavior across all implementations.

## Decision
### Token Bucket
- **State**: `{ tokens: float, last_refill_ts: int64 }`
- **Logic**: On allow, atomically check if `tokens >= 1`. If yes, decrement and return allowed. Otherwise, calculate `retry_at = last_refill_ts + (1 / refill_rate)`, return retry_after.
- **Refill**: Always uses server-side `Redis TIME` for refill calculations to avoid clock skew.

### Sliding Window
- **State**: Sorted set with `member=timestamp` and `score=timestamp` (milliseconds)
- **Logic**: Remove entries older than `window_size`. Count existing entries. If `count < limit`, add current timestamp, return allowed. Otherwise, return oldest timestamp as retry_after.
- **Memory bound**: Cap sorted set size at `limit * 2` entries to prevent unbounded growth.

### Leaky Bucket
- **State**: `{ fill_level: float, last_update_ts: int64 }`
- **Logic**: On allow, calculate drained amount via `drain_rate * (now - last_update_ts)`, subtract from fill_level (floor at 0). If `fill_level < capacity`, add 1 to fill_level and return allowed. Return retry_after based on time to empty one slot.

### Shared Constraints
- All algorithms use **Redis server-side timestamp** (`redis.call('TIME')`) as the time source.
- All state mutations happen within a single Lua script for atomicity.
- Responses include: `allowed` (bool), `remaining` (int), `retry_after_ms` (int), `reset_at_ms` (int).

## Consequences
- Clock skew issues mitigated by server-side time.
- Memory bounded for Sliding Window implementation.
- Consistent response schema across all algorithms enables uniform client code.