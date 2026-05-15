package ratelimiter

import (
	"context"
	"errors"
	"testing"
)

type mockRedis struct {
	loadedScript string
	keys         []string
	args         []any
	result       any
}

func (m *mockRedis) ScriptLoad(_ context.Context, script string) (string, error) {
	m.loadedScript = script
	return "sha", nil
}

func (m *mockRedis) EvalSHA(_ context.Context, _ string, keys []string, args ...any) (any, error) {
	m.keys = keys
	m.args = args
	return m.result, nil
}

func TestRateLimiterCheck(t *testing.T) {
	redis := &mockRedis{result: []any{1, 4, 0, 100}}
	limiter, err := New(redis, WithScriptSource("return {1,4,0,100}"), WithKeyPrefix("test"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	decision, err := limiter.Check(context.Background(), "user:1", TokenBucketPolicy{Capacity: 5, RefillRate: 1, TTLSec: 60}, 1)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !decision.Allowed || decision.Remaining != 4 {
		t.Fatalf("decision = %#v", decision)
	}
	if got, want := redis.keys[0], "test:token_bucket:key:dXNlcjox"; got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
	wantArgs := []any{int64(5), float64(1), int64(1), int64(60)}
	for i := range wantArgs {
		if redis.args[i] != wantArgs[i] {
			t.Fatalf("arg[%d] = %#v, want %#v", i, redis.args[i], wantArgs[i])
		}
	}
}

func TestRateLimiterDefaultPrefixUsesLogicalPrefix(t *testing.T) {
	redis := &mockRedis{result: []any{1, 1, 0, 100}}
	limiter, err := New(redis, WithScriptSource("return {1,1,0,100}"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = limiter.Check(context.Background(), "user:1", TokenBucketPolicy{Capacity: 1, RefillRate: 1, TTLSec: 60}, 1)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if got, want := redis.keys[0], "ratelimit:token_bucket:key:dXNlcjox"; got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

func TestRateLimiterNormalizesLegacyAlgorithmPrefix(t *testing.T) {
	redis := &mockRedis{result: []any{1, 1, 0, 100}}
	limiter, err := New(redis, WithScriptSource("return {1,1,0,100}"), WithKeyPrefix("ratelimit:token_bucket"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = limiter.Check(context.Background(), "user:1", TokenBucketPolicy{Capacity: 1, RefillRate: 1, TTLSec: 60}, 1)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if got, want := redis.keys[0], "ratelimit:token_bucket:key:dXNlcjox"; got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

func TestRateLimiterInvalidPolicy(t *testing.T) {
	limiter, err := New(&mockRedis{}, WithScriptSource("return {1,1,0,0}"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = limiter.Check(context.Background(), "key", TokenBucketPolicy{Capacity: 0, RefillRate: 1, TTLSec: 60}, 1)
	if !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("error = %v, want ErrInvalidPolicy", err)
	}
}

func TestRegistryUnknownAlgorithm(t *testing.T) {
	_, err := Registry.Create("missing")
	if !errors.Is(err, ErrUnknownAlgorithm) {
		t.Fatalf("error = %v, want ErrUnknownAlgorithm", err)
	}
}

func TestCreateRateLimiterByName(t *testing.T) {
	_, err := CreateRateLimiter(&mockRedis{}, "token_bucket", WithScriptSource("return {1,1,0,0}"))
	if err != nil {
		t.Fatalf("CreateRateLimiter() error = %v", err)
	}
}
