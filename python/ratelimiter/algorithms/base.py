from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any, Protocol

from ..models import Decision


@dataclass(frozen=True)
class BasePolicy:
    """Marker base class for algorithm policies."""


class RateLimitAlgorithm(Protocol):
    """Algorithm contract used by RateLimiter orchestration."""

    name: str
    key_prefix: str
    policy_type: type[BasePolicy]

    def default_script_path(self) -> Path:
        ...

    def validate_policy(self, policy: BasePolicy) -> None:
        ...

    def build_redis_args(self, policy: BasePolicy, requested: int) -> list[Any]:
        ...

    def build_redis_keys(self, prefix: str | None, key: str) -> list[str]:
        ...
    
    def parse_decision(self, raw: Any) -> Decision:
        ...
