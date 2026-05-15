package algorithms

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// TokenBucketPolicy configures the token-bucket algorithm.
type TokenBucketPolicy struct {
	Capacity   int64
	RefillRate float64
	TTLSec     int64
}

// Validate verifies token-bucket policy invariants.
func (p TokenBucketPolicy) Validate() error {
	if p.Capacity <= 0 {
		return fmt.Errorf("capacity must be > 0")
	}
	if p.RefillRate < 0 {
		return fmt.Errorf("refill_rate must be >= 0")
	}
	if p.TTLSec <= 0 {
		return fmt.Errorf("ttl_sec must be > 0")
	}
	return nil
}

// TokenBucketAlgorithm implements the Redis Lua token-bucket contract.
type TokenBucketAlgorithm struct{}

func (TokenBucketAlgorithm) Name() string { return "token_bucket" }

func (TokenBucketAlgorithm) KeyPrefix() string { return "ratelimit" }

func (TokenBucketAlgorithm) BuildRedisKeys(baseKey string) []string { return []string{baseKey} }

func (TokenBucketAlgorithm) UseRedisHashTag() bool { return false }

func (TokenBucketAlgorithm) DefaultScriptPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("..", "..", "..", "scripts", "lua", "token_bucket.lua")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "scripts", "lua", "token_bucket.lua"))
}

func (TokenBucketAlgorithm) ValidatePolicy(policy any) error {
	p, ok := asTokenBucketPolicy(policy)
	if !ok {
		return fmt.Errorf("token_bucket algorithm requires TokenBucketPolicy")
	}
	return p.Validate()
}

func (TokenBucketAlgorithm) BuildRedisArgs(policy any, requested int64) ([]any, error) {
	p, ok := asTokenBucketPolicy(policy)
	if !ok {
		return nil, fmt.Errorf("token_bucket algorithm requires TokenBucketPolicy")
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return []any{p.Capacity, p.RefillRate, requested, p.TTLSec}, nil
}

func (TokenBucketAlgorithm) ParseDecision(raw any) (Decision, error) {
	return parseNormalizedDecision(raw)
}

func asTokenBucketPolicy(policy any) (TokenBucketPolicy, bool) {
	switch p := policy.(type) {
	case TokenBucketPolicy:
		return p, true
	case *TokenBucketPolicy:
		if p == nil {
			return TokenBucketPolicy{}, false
		}
		return *p, true
	default:
		return TokenBucketPolicy{}, false
	}
}
