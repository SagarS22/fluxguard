from dataclasses import dataclass
from pathlib import Path
from typing import Any

from ..models import Decision, ScriptExecutionError
from .base import BasePolicy, RateLimitAlgorithm


@dataclass(frozen=True)
class SlidingWindowPolicy(BasePolicy):
    capacity: int
    window_size: int
    ttl_sec: int

    def validate(self) -> None:
        if self.capacity <= 0:
            raise ValueError("capacity must be > 0")
        if self.window_size < 0:
            raise ValueError("window size must be >= 0")
        if self.ttl_sec <= 0:
            raise ValueError("ttl_sec must be > 0")


class SlidingWindowAlgorithm(RateLimitAlgorithm):
    name = "sliding_window"
    key_prefix = "ratelimit:sliding_window"
    policy_type = SlidingWindowPolicy

    def default_script_path(self) -> Path:
        return Path(__file__).resolve().parents[3] / "scripts" / "lua" / "sliding_window.lua"

    def validate_policy(self, policy: BasePolicy) -> None:
        if not isinstance(policy, SlidingWindowPolicy):
            raise TypeError("sliding_window algorithm requires SlidingWindowPolicy")
        policy.validate()

    def build_redis_args(self, policy: BasePolicy, requested: int) -> list[Any]:
        assert isinstance(policy, SlidingWindowPolicy)
        return [policy.capacity, policy.window_size, requested, policy.ttl_sec]

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