from ratelimiter import Policy, RateLimiter, ScriptExecutionError


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
