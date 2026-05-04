from pathlib import Path

from ratelimiter import (
    Policy,
    RateLimiter,
    ScriptExecutionError,
    TokenBucketAlgorithm,
    create_rate_limiter,
    registry,
)


class FakeRedis:
    def __init__(self, *, response=None, fail_once_noscript=False, fail_error=None):
        self.response = response or [1, 4, 0, 123]
        self.fail_once_noscript = fail_once_noscript
        self.fail_error = fail_error
        self.loaded = 0

    def script_load(self, script):
        self.loaded += 1
        return f"sha-{self.loaded}"

    def evalsha(self, sha, numkeys, key, *args):
        if self.fail_once_noscript:
            self.fail_once_noscript = False
            raise Exception("NOSCRIPT No matching script")
        if self.fail_error:
            raise Exception(self.fail_error)
        return self.response


def test_policy_validation():
    p = Policy(capacity=1, refill_rate=1.0, ttl_sec=1)
    p.validate()


def test_check_maps_response():
    rl = RateLimiter(FakeRedis())
    d = rl.check(key="user:1", policy=Policy(capacity=5, refill_rate=1.0, ttl_sec=30), requested=1)
    assert d.allowed is True
    assert d.remaining == 4


def test_check_noscript_reloads():
    redis = FakeRedis(fail_once_noscript=True)
    rl = RateLimiter(redis)
    d = rl.check(key="user:1", policy=Policy(capacity=5, refill_rate=1.0, ttl_sec=30), requested=1)
    assert d.allowed is True
    assert redis.loaded >= 2


def test_check_raises_on_non_noscript_error():
    rl = RateLimiter(FakeRedis(fail_error="connection lost"))
    try:
        rl.check(key="user:1", policy=Policy(capacity=5, refill_rate=1.0, ttl_sec=30), requested=1)
        assert False, "expected ScriptExecutionError"
    except ScriptExecutionError:
        assert True


def test_factory_by_name_and_instance():
    redis = FakeRedis()
    rl1 = create_rate_limiter(redis, algorithm="token_bucket")
    rl2 = create_rate_limiter(redis, algorithm=TokenBucketAlgorithm())
    policy = Policy(capacity=1, refill_rate=1.0, ttl_sec=1)
    assert rl1.check(key="k1", policy=policy).allowed
    assert rl2.check(key="k2", policy=policy).allowed


def test_registry_unknown_algorithm_error():
    try:
        create_rate_limiter(FakeRedis(), algorithm="missing")
        assert False, "expected ValueError"
    except ValueError as exc:
        assert "unknown algorithm" in str(exc)


def test_default_script_path_exists():
    path = TokenBucketAlgorithm().default_script_path()
    assert Path(path).exists()


def test_policy_alias_backward_compatibility():
    p = Policy(capacity=2, refill_rate=1.0, ttl_sec=10)
    assert p.capacity == 2


def test_registry_registration_lifecycle():
    class CustomAlgo(TokenBucketAlgorithm):
        name = "custom"

    registry.register("custom", CustomAlgo)
    rl = create_rate_limiter(FakeRedis(), algorithm="custom")
    assert rl.check(key="x", policy=Policy(capacity=1, refill_rate=1.0, ttl_sec=1)).allowed
    registry.unregister("custom")
