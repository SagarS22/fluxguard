# ADR 0002: API Model

## Status
**Accepted** | Sprint 1

## Context
We need a stable API model that works across all three algorithms and can be implemented consistently in Go, Python, and the sidecar service. The API must support both simple use cases and advanced scenarios requiring retry information.

## Decision
### Core API Methods
```go
// Allow - simple pass/fail check
Allow(ctx context.Context, key string, policy Policy) (Result, error)

// Reserve - advanced check with retry information
Reserve(ctx context.Context, key string, policy Policy) (Result, error)

// Reset - manually reset rate limit state
Reset(ctx context.Context, key string) error
```

### Policy Definition
```go
type Policy struct {
    Limit        int     // Max requests allowed in window
    Burst       int     // Initial burst capacity (for token bucket)
    RefillRate  float64 // Tokens/requests per second (token bucket)
    WindowMs   int64   // Window size in milliseconds (sliding window)
    Capacity   int     // Bucket capacity (leaky bucket)
    DrainRate  float64 // Drain rate per second (leaky bucket)
}
```

### Result Schema
```go
type Result struct {
    Allowed       bool  // Whether request is allowed
    Remaining     int   // Remaining requests in window
    RetryAfterMs int64 // Milliseconds until next allow (if not allowed)
    ResetAtMs     int64 // Timestamp when bucket resets
}
```

### Algorithm Selection
- Algorithm inferred from Policy fields:
  - `RefillRate > 0` → Token Bucket
  - `WindowMs > 0 && RefillRate == 0` → Sliding Window
  - `Capacity > 0 && DrainRate > 0` → Leaky Bucket

## Consequences
- Single API surface across Go, Python, and service.
- Extensible: new algorithms can be added by inferring from policy fields.
- Backward compatible: adding new fields to Policy is non-breaking.