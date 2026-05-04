from __future__ import annotations

from typing import Any

from .algorithms.base import RateLimitAlgorithm
from .ratelimiter import RateLimiter
from .registry import registry


def create_rate_limiter(
    redis_client: Any,
    *,
    algorithm: str | RateLimitAlgorithm = "token_bucket",
    script_path: str | None = None,
    key_prefix: str | None = None,
) -> RateLimiter:
    algo = registry.create(algorithm) if isinstance(algorithm, str) else algorithm
    return RateLimiter(redis_client, algorithm=algo, script_path=script_path, key_prefix=key_prefix)
