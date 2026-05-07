from pathlib import Path

import pytest

from ratelimiter import RateLimiter, SlidingWindowAlgorithm, SlidingWindowPolicy, create_rate_limiter


class FakeRedis:
    def __init__(self, *, response=None, fail_once_noscript=False):
        self.response = response or [1, 2, 0, 123]
        self.fail_once_noscript = fail_once_noscript
        self.loaded = 0
        self.calls = []

    def script_load(self, script):
        self.loaded += 1
        return f"sha-{self.loaded}"

    def evalsha(self, sha, numkeys, *keys_and_args):
        self.calls.append((sha, numkeys, keys_and_args))
        if self.fail_once_noscript:
            self.fail_once_noscript = False
            raise Exception("NOSCRIPT No matching script")
        return self.response


def test_sliding_window_passes_two_redis_keys_and_converts_window_sec_to_ms():
    redis = FakeRedis()
    rl = create_rate_limiter(redis, algorithm="sliding_window")
    policy = SlidingWindowPolicy(capacity=3, window_sec=60, ttl_sec=30)

    decision = rl.check(key="user:1", policy=policy, requested=1)

    assert decision.allowed is True
    assert redis.calls == [
        (
            "sha-1",
            2,
            (
                "ratelimit:sliding_window:user:1",
                "ratelimit:sliding_window:user:1:seq",
                3,
                60_000,
                1,
                30,
            ),
        )
    ]


def test_sliding_window_noscript_retry_keeps_two_key_count():
    redis = FakeRedis(fail_once_noscript=True)
    rl = RateLimiter(redis, algorithm=SlidingWindowAlgorithm())
    policy = SlidingWindowPolicy(capacity=3, window_sec=1, ttl_sec=30)

    decision = rl.check(key="user:1", policy=policy, requested=2)

    assert decision.allowed is True
    assert redis.loaded == 2
    assert [call[1] for call in redis.calls] == [2, 2]
    assert redis.calls[1][2] == (
        "ratelimit:sliding_window:user:1",
        "ratelimit:sliding_window:user:1:seq",
        3,
        1_000,
        2,
        30,
    )


def test_sliding_window_accepts_subsecond_windows_as_milliseconds():
    redis = FakeRedis()
    rl = RateLimiter(redis, algorithm=SlidingWindowAlgorithm())
    policy = SlidingWindowPolicy(capacity=3, window_sec=0.25, ttl_sec=30)

    rl.check(key="user:1", policy=policy)

    assert redis.calls[0][2][3] == 250


def test_sliding_window_rejects_zero_window_before_lua_execution():
    redis = FakeRedis()
    rl = RateLimiter(redis, algorithm=SlidingWindowAlgorithm())
    policy = SlidingWindowPolicy(capacity=3, window_sec=0, ttl_sec=30)

    with pytest.raises(ValueError, match="window_sec must be > 0"):
        rl.check(key="user:1", policy=policy)

    assert redis.calls == []


def test_sliding_window_default_script_path_exists():
    assert Path(SlidingWindowAlgorithm().default_script_path()).exists()
