from __future__ import annotations

from pathlib import Path
from typing import Any

from .algorithms.base import BasePolicy, RateLimitAlgorithm
from .algorithms.token_bucket import TokenBucketAlgorithm
from .keys import RedisKeyBuilder
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
        self._key_prefix = (key_prefix or "ratelimit").strip(":")
        self._key_builder = RedisKeyBuilder()
        self._executor = RedisScriptExecutor(
            redis_client,
            script_path=script_path or self._algorithm.default_script_path(),
        )

    @property
    def policy_type(self) -> type[BasePolicy]:
        return self._algorithm.policy_type

    def check(self, *, key: str, policy: BasePolicy, requested: int = 1, dimension: str = "key") -> Decision:
        if not key:
            raise ValueError("key must be non-empty")
        if requested < 0:
            raise ValueError("requested must be >= 0")

        self._algorithm.validate_policy(policy)

        base_key = self._key_builder.build_base_key(
            prefix=self._key_prefix,
            algorithm=self._algorithm.name,
            dimension=dimension,
            key=key,
            hash_tag=self._algorithm.use_redis_hash_tag,
        )
        redis_keys = self._algorithm.build_redis_keys(base_key)
        for redis_key in redis_keys:
            self._key_builder.validate_key_length(redis_key)

        args = self._algorithm.build_redis_args(policy, requested)
        raw = self._executor.execute(redis_keys=redis_keys, args=args)
        return self._algorithm.parse_decision(raw)

