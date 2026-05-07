from ratelimiter import RateLimiter, TokenBucketPolicy
from ratelimiter.algorithms.base import BasePolicy
from ratelimiter.models import Decision


class FakeRedis:
    def script_load(self, script):
        return "sha"

    def evalsha(self, sha, numkeys, *keys_and_args):
        return [1, 9, 0, 111]


class FakeAlgorithm:
    name = "fake"
    policy_type = TokenBucketPolicy
    use_redis_hash_tag = False

    def __init__(self):
        self.validated = False
        self.args_built = False
        self.keys_built = False
        self.parsed = False

    def default_script_path(self):
        from pathlib import Path

        return Path(__file__).resolve().parents[2] / "scripts" / "lua" / "token_bucket.lua"

    def validate_policy(self, policy: BasePolicy):
        self.validated = True

    def build_redis_args(self, policy: BasePolicy, requested: int):
        self.args_built = True
        return [1, 1.0, requested, 30]

    def build_redis_keys(self, base_key: str):
        self.keys_built = True
        return [base_key]

    def parse_decision(self, raw):
        self.parsed = True
        return Decision(True, 9, 0, 111)


def test_ratelimiter_uses_algorithm_inversion_of_control():
    algo = FakeAlgorithm()
    rl = RateLimiter(FakeRedis(), algorithm=algo)
    d = rl.check(key="user:1", policy=TokenBucketPolicy(capacity=1, refill_rate=1.0, ttl_sec=30), requested=1)
    assert d.allowed is True
    assert algo.validated is True
    assert algo.args_built is True
    assert algo.keys_built is True
    assert algo.parsed is True
