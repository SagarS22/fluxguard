package ratelimiter

import (
	"context"
	"fmt"
	"strings"

	"github.com/SagarS22/fluxguard/go/ratelimiter/algorithms"
	"github.com/SagarS22/fluxguard/go/ratelimiter/internal/script"
)

// RedisClient is the minimal Redis scripting API required by the SDK.
type RedisClient = script.RedisClient

// RateLimiter orchestrates policy validation, Redis script execution, and response parsing.
type RateLimiter struct {
	algorithm Algorithm
	keyPrefix string
	executor  *script.Executor
}

// New constructs a RateLimiter with token_bucket as the default algorithm.
func New(client RedisClient, opts ...Option) (*RateLimiter, error) {
	cfg := config{algorithm: algorithms.TokenBucketAlgorithm{}}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	if cfg.algorithm == nil {
		return nil, fmt.Errorf("%w: nil algorithm", ErrUnknownAlgorithm)
	}
	keyPrefix := strings.Trim(cfg.algorithm.KeyPrefix(), ":")
	if cfg.keyPrefix != "" {
		keyPrefix = cfg.keyPrefix
	}
	if keyPrefix == "" {
		return nil, fmt.Errorf("key prefix must be non-empty")
	}

	scriptPath := cfg.scriptPath
	if scriptPath == "" && cfg.scriptSource == "" {
		scriptPath = cfg.algorithm.DefaultScriptPath()
	}
	executor, err := script.NewExecutor(client, scriptPath, cfg.scriptSource)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScriptExecution, err)
	}

	return &RateLimiter{algorithm: cfg.algorithm, keyPrefix: keyPrefix, executor: executor}, nil
}

// Check evaluates one rate-limit decision for key and policy.
func (r *RateLimiter) Check(ctx context.Context, key string, policy any, requested int64) (Decision, error) {
	if r == nil {
		return Decision{}, fmt.Errorf("rate limiter is nil")
	}
	if strings.TrimSpace(key) == "" {
		return Decision{}, fmt.Errorf("key must be non-empty")
	}
	if requested < 0 {
		return Decision{}, fmt.Errorf("requested must be >= 0")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.algorithm.ValidatePolicy(policy); err != nil {
		return Decision{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}
	args, err := r.algorithm.BuildRedisArgs(policy, requested)
	if err != nil {
		return Decision{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}

	raw, err := r.executor.Execute(ctx, r.keyPrefix+":"+key, args)
	if err != nil {
		return Decision{}, fmt.Errorf("%w: %v", ErrScriptExecution, err)
	}
	decision, err := r.algorithm.ParseDecision(raw)
	if err != nil {
		return Decision{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	return decision, nil
}
