from __future__ import annotations

from pathlib import Path
from typing import Any

from .models import ScriptExecutionError


class RedisScriptExecutor:
    def __init__(self, redis_client: Any, *, script_path: str | Path | None = None, script_source: str | None = None) -> None:
        self._redis = redis_client
        self._script_sha: str | None = None
        self._script = self._load_script_source(script_path=script_path, script_source=script_source)

    @staticmethod
    def _load_script_source(*, script_path: str | Path | None, script_source: str | None) -> str:
        if script_source is not None:
            content = script_source
        elif script_path is not None:
            content = Path(script_path).read_text(encoding="utf-8")
        else:
            raise ValueError("script_path or script_source must be provided")

        if not content.strip():
            raise ValueError("Lua script is empty")
        return content

    def _ensure_script_loaded(self) -> str:
        if self._script_sha is None:
            self._script_sha = self._redis.script_load(self._script)
        return self._script_sha

    def execute(self, *, redis_key: str, args: list[Any]) -> Any:
        sha = self._ensure_script_loaded()
        try:
            return self._redis.evalsha(sha, 1, redis_key, *args)
        except Exception as exc:
            if "NOSCRIPT" in str(exc):
                try:
                    self._script_sha = self._redis.script_load(self._script)
                    return self._redis.evalsha(self._script_sha, 1, redis_key, *args)
                except Exception as reload_exc:
                    raise ScriptExecutionError(f"redis script reload/execution failed: {reload_exc}") from reload_exc
            raise ScriptExecutionError(f"redis script execution failed: {exc}") from exc
