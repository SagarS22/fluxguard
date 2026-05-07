import pytest

from ratelimiter import RateLimiter, TokenBucketPolicy
from ratelimiter.keys import RedisKeyBuilder, encode_key_component


class FakeRedis:
    def __init__(self):
        self.calls = []

    def script_load(self, script):
        return "sha"

    def evalsha(self, sha, numkeys, *keys_and_args):
        self.calls.append((sha, numkeys, keys_and_args))
        return [1, 4, 0, 123]


def test_encode_key_component_uses_urlsafe_base64_without_padding():
    assert encode_key_component("user:123") == "dXNlcjoxMjM"


def test_key_builder_uses_adr_shape_and_encodes_user_key():
    key = RedisKeyBuilder().build_base_key(
        prefix="ratelimit",
        algorithm="token_bucket",
        dimension="user",
        key="user:123",
    )

    assert key == "ratelimit:token_bucket:user:dXNlcjoxMjM"
    assert "user:123" not in key


def test_key_builder_can_hash_tag_encoded_key_for_redis_cluster():
    key = RedisKeyBuilder().build_base_key(
        prefix="ratelimit",
        algorithm="sliding_window",
        dimension="user",
        key="user:123",
        hash_tag=True,
    )

    assert key == "ratelimit:sliding_window:user:{dXNlcjoxMjM}"


def test_key_builder_rejects_invalid_structural_components():
    builder = RedisKeyBuilder()

    with pytest.raises(ValueError, match="dimension may only contain"):
        builder.build_base_key(prefix="ratelimit", algorithm="token_bucket", dimension="user:id", key="123")

    with pytest.raises(ValueError, match="prefix may only contain"):
        builder.build_base_key(prefix="rate:limit", algorithm="token_bucket", dimension="user", key="123")


def test_key_builder_rejects_oversized_keys():
    builder = RedisKeyBuilder(max_key_bytes=40)

    with pytest.raises(ValueError, match="redis key exceeds 40 bytes"):
        builder.build_base_key(prefix="ratelimit", algorithm="token_bucket", dimension="user", key="x" * 100)


def test_rate_limiter_uses_safe_encoded_key_and_dimension():
    redis = FakeRedis()
    rl = RateLimiter(redis)
    policy = TokenBucketPolicy(capacity=5, refill_rate=1.0, ttl_sec=30)

    rl.check(key="user:123", dimension="user", policy=policy)

    assert redis.calls[0] == (
        "sha",
        1,
        ("ratelimit:token_bucket:user:dXNlcjoxMjM", 5, 1.0, 1, 30),
    )


def test_rate_limiter_rejects_oversized_keys_before_redis_execution():
    redis = FakeRedis()
    rl = RateLimiter(redis)
    policy = TokenBucketPolicy(capacity=5, refill_rate=1.0, ttl_sec=30)

    with pytest.raises(ValueError, match="redis key exceeds 512 bytes"):
        rl.check(key="x" * 400, dimension="user", policy=policy)

    assert redis.calls == []
