from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from ..models import Decision, ScriptExecutionError
from .base import BasePolicy, RateLimitAlgorithm


@dataclass(frozen=True)
class TokenBucketPolicy(BasePolicy):
    capacity: int
    refill_rate: float
    ttl_sec: int = 60

    def validate(self) -> None:
        if self.capacity <= 0:
            raise ValueError("capacity must be > 0")
        if self.refill_rate < 0:
            raise ValueError("refill_rate must be >= 0")
        if self.ttl_sec <= 0:
            raise ValueError("ttl_sec must be > 0")


class TokenBucketAlgorithm(RateLimitAlgorithm):
    name = "token_bucket"
    key_prefix = "ratelimit:token_bucket"
    policy_type = TokenBucketPolicy

    def default_script_path(self) -> Path:
        return Path(__file__).resolve().parents[3] / "scripts" / "lua" / "token_bucket.lua"

    def validate_policy(self, policy: BasePolicy) -> None:
        if not isinstance(policy, TokenBucketPolicy):
            raise TypeError("token_bucket algorithm requires TokenBucketPolicy")
        policy.validate()

    def build_redis_args(self, policy: BasePolicy, requested: int) -> list[Any]:
        assert isinstance(policy, TokenBucketPolicy)
        return [policy.capacity, policy.refill_rate, requested, policy.ttl_sec]

    def parse_decision(self, raw: Any) -> Decision:
        if not isinstance(raw, (list, tuple)) or len(raw) != 4:
            raise ScriptExecutionError(f"unexpected script response: {raw!r}")

        allowed, remaining, retry_after_ms, reset_at_ms = raw
        return Decision(
            allowed=bool(int(allowed)),
            remaining=int(remaining),
            retry_after_ms=int(retry_after_ms),
            reset_at_ms=int(reset_at_ms),
        )
