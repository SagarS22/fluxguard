import os
import time

import pytest

from ratelimiter import Policy, RateLimiter


redis = pytest.importorskip("redis")


@pytest.fixture
def redis_client():
    url = os.getenv("REDIS_URL", "redis://localhost:6379/0")
    client = redis.Redis.from_url(url, decode_responses=False)
    try:
        client.ping()
    except Exception as exc:
        pytest.skip(f"redis unavailable: {exc}")
    yield client


def test_allow_then_deny_then_refill(redis_client):
    key = f"it:user:{time.time_ns()}"
    rl = RateLimiter(redis_client)
    policy = Policy(capacity=2, refill_rate=2.0, ttl_sec=30)

    first = rl.check(key=key, policy=policy, requested=1)
    second = rl.check(key=key, policy=policy, requested=1)
    third = rl.check(key=key, policy=policy, requested=1)

    assert first.allowed is True
    assert second.allowed is True
    assert third.allowed is False
    assert third.retry_after_ms > 0

    time.sleep(0.6)
    fourth = rl.check(key=key, policy=policy, requested=1)
    assert fourth.allowed is True
