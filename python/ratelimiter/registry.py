from __future__ import annotations

from collections.abc import Callable

from .algorithms.base import RateLimitAlgorithm
from .algorithms.token_bucket import TokenBucketAlgorithm

AlgorithmFactory = Callable[[], RateLimitAlgorithm]


class AlgorithmRegistry:
    def __init__(self) -> None:
        self._factories: dict[str, AlgorithmFactory] = {
            "token_bucket": TokenBucketAlgorithm,
        }

    def register(self, name: str, factory: AlgorithmFactory) -> None:
        if not name:
            raise ValueError("algorithm name must be non-empty")
        self._factories[name] = factory

    def unregister(self, name: str) -> None:
        self._factories.pop(name, None)

    def create(self, name: str) -> RateLimitAlgorithm:
        try:
            return self._factories[name]()
        except KeyError as exc:
            raise ValueError(f"unknown algorithm: {name}") from exc


registry = AlgorithmRegistry()
