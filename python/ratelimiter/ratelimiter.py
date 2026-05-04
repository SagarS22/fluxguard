from __future__ import annotations

from pathlib import Path
from typing import Any

from .algorithms.base import BasePolicy, RateLimitAlgorithm
from .algorithms.token_bucket import TokenBucketAlgorithm, TokenBucketPolicy
from .models import Decision
from .script_executor import RedisScriptExecutor


class RateLimiter:
    def __init__(
        self,
        redis_client: Any,
        *,
        algorithm: RateLimitAlgorithm | None = None,
        script_path: str | Path | None = None,
        key_prefix: str | None = None,
    ) -> None:
        self._algorithm = algorithm or TokenBucketAlgorithm()
        self._key_prefix = (key_prefix or self._algorithm.key_prefix).strip(":")
        self._executor = RedisScriptExecutor(
            redis_client,
            script_path=script_path or self._algorithm.default_script_path(),
        )

    def check(self, *, key: str, policy: BasePolicy, requested: int = 1) -> Decision:
        if not key:
            raise ValueError("key must be non-empty")
        if requested < 0:
            raise ValueError("requested must be >= 0")

        self._algorithm.validate_policy(policy)

        redis_key = f"{self._key_prefix}:{key}"
        args = self._algorithm.build_redis_args(policy, requested)
        raw = self._executor.execute(redis_key=redis_key, args=args)
        return self._algorithm.parse_decision(raw)


# Backward compatibility alias
Policy = TokenBucketPolicy
