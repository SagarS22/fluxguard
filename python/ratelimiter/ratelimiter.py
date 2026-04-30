from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any


class RateLimiterError(Exception):
    """Base exception for SDK errors."""


class ScriptExecutionError(RateLimiterError):
    """Raised when Redis script execution fails."""


@dataclass(frozen=True)
class Policy:
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


@dataclass(frozen=True)
class Decision:
    allowed: bool
    remaining: int
    retry_after_ms: int
    reset_at_ms: int


class RateLimiter:
    def __init__(
        self,
        redis_client: Any,
        *,
        script_path: str | Path | None = None,
        key_prefix: str = "ratelimit:token_bucket",
    ) -> None:
        self._redis = redis_client
        self._script_sha: str | None = None
        self._key_prefix = key_prefix.strip(":")
        self._script = self._load_script_source(script_path)

    @staticmethod
    def _load_script_source(script_path: str | Path | None) -> str:
        if script_path is None:
            script_path = (
                Path(__file__).resolve().parents[2] / "scripts" / "lua" / "token_bucket.lua"
            )
        content = Path(script_path).read_text(encoding="utf-8")
        if not content.strip():
            raise ValueError("Lua script is empty")
        return content

    def _ensure_script_loaded(self) -> str:
        if self._script_sha is None:
            self._script_sha = self._redis.script_load(self._script)
        return self._script_sha

    def check(self, *, key: str, policy: Policy, requested: int = 1) -> Decision:
        policy.validate()
        if not key:
            raise ValueError("key must be non-empty")
        if requested < 0:
            raise ValueError("requested must be >= 0")

        redis_key = f"{self._key_prefix}:{key}"
        args = [policy.capacity, policy.refill_rate, requested, policy.ttl_sec]
        sha = self._ensure_script_loaded()

        try:
            raw = self._redis.evalsha(sha, 1, redis_key, *args)
        except Exception as exc:
            if "NOSCRIPT" in str(exc):
                try:
                    self._script_sha = self._redis.script_load(self._script)
                    raw = self._redis.evalsha(self._script_sha, 1, redis_key, *args)
                except Exception as reload_exc:
                    raise ScriptExecutionError(f"redis script reload/execution failed: {reload_exc}") from reload_exc
            else:
                raise ScriptExecutionError(f"redis script execution failed: {exc}") from exc

        if not isinstance(raw, (list, tuple)) or len(raw) != 4:
            raise ScriptExecutionError(f"unexpected script response: {raw!r}")

        allowed, remaining, retry_after_ms, reset_at_ms = raw
        return Decision(
            allowed=bool(int(allowed)),
            remaining=int(remaining),
            retry_after_ms=int(retry_after_ms),
            reset_at_ms=int(reset_at_ms),
        )
