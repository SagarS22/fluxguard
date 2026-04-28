# ADR 0003: Redis Key Strategy

## Status
**Accepted** | Sprint 1

## Context
We need a consistent key naming strategy that supports multi-tenancy, prevents collisions, enables keyspace monitoring, and allows TTL management across all algorithms.

## Decision
### Key Format
```
{prefix}:{algorithm}:{dimension}:{key}
```

### Components
| Component | Description | Example |
|----------|-------------|----------|
| prefix | Application prefix | `ratelimit` |
| algorithm | Algorithm identifier | `token_bucket`, `sliding_window`, `leaky_bucket` |
| dimension | Rate limit dimension | `user`, `ip`, `api_key` |
| key | The actual key value | `user:123`, `192.168.1.1` |

### Examples
```
ratelimit:token_bucket:user:user:123
ratelimit:sliding_window:ip:192.168.1.1
ratelimit:leaky_bucket:api_key:abc123
```

### TTL Strategy
- **Default TTL**: Max of `policy.WindowMs` (converted to seconds) or 3600 seconds, plus 60s buffer.
- **Auto-extension**: On each allow, TTL is refreshed to the maximum (prevents key expiry during active use).
- **Expiration**: Keys expire naturally if no activity for TTL period.

### Key Encoding
- Colons (`:`) reserved as delimiter; keys with colons are encoded via URL-safe base64.
- Redis key max length: 512 bytes (enforced at registration).

### Script SHA Caching
- Scripts loaded once on startup; SHA1 digest used for subsequent calls.
- Fallback: If SHA mismatch (Redis restart), reload script and cache new SHA.

## Consequences
- Clear key hierarchy enables monitoring per algorithm/dimension.
- TTL strategy prevents stale keys while allowing natural cleanup.
- SHA caching improves performance and provides Redis restart resilience.