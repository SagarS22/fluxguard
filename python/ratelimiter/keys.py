from __future__ import annotations

import base64
import re
from dataclasses import dataclass

_STRUCTURAL_COMPONENT_RE = re.compile(r"^[A-Za-z0-9_-]+$")


def encode_key_component(value: str) -> str:
    """Encode a user-controlled key component for safe colon-delimited Redis keys."""
    if not value:
        raise ValueError("key must be non-empty")
    encoded = base64.urlsafe_b64encode(value.encode("utf-8")).decode("ascii")
    return encoded.rstrip("=")


@dataclass(frozen=True)
class RedisKeyBuilder:
    """Build ADR-compliant Redis keys for rate-limiter scripts."""

    max_key_bytes: int = 512

    def build_base_key(self, *, prefix: str, algorithm: str, dimension: str, key: str, hash_tag: bool = False) -> str:
        prefix = self._validate_structural_component("prefix", prefix)
        algorithm = self._validate_structural_component("algorithm", algorithm)
        dimension = self._validate_structural_component("dimension", dimension)
        encoded_key = encode_key_component(key)
        key_component = f"{{{encoded_key}}}" if hash_tag else encoded_key

        redis_key = f"{prefix}:{algorithm}:{dimension}:{key_component}"
        self.validate_key_length(redis_key)
        return redis_key

    def validate_key_length(self, redis_key: str) -> None:
        key_bytes = len(redis_key.encode("utf-8"))
        if key_bytes > self.max_key_bytes:
            raise ValueError(f"redis key exceeds {self.max_key_bytes} bytes")

    @staticmethod
    def _validate_structural_component(name: str, value: str) -> str:
        value = value.strip(":") if value else ""
        if not value:
            raise ValueError(f"{name} must be non-empty")
        if not _STRUCTURAL_COMPONENT_RE.fullmatch(value):
            raise ValueError(f"{name} may only contain letters, numbers, '_' and '-'")
        return value
