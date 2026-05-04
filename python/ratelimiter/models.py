from __future__ import annotations

from dataclasses import dataclass


class RateLimiterError(Exception):
    """Base exception for SDK errors."""


class ScriptExecutionError(RateLimiterError):
    """Raised when Redis script execution fails."""


@dataclass(frozen=True)
class Decision:
    allowed: bool
    remaining: int
    retry_after_ms: int
    reset_at_ms: int
