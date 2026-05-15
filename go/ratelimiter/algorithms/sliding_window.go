package algorithms

import (
	"fmt"
	"math"
	"path/filepath"
	"runtime"
)

// SlidingWindowPolicy configures the sliding-window-log algorithm.
type SlidingWindowPolicy struct {
	Capacity  int64
	WindowSec float64
	TTLSec    int64
}

// Validate verifies sliding-window policy invariants.
func (p SlidingWindowPolicy) Validate() error {
	if p.Capacity <= 0 {
		return fmt.Errorf("capacity must be > 0")
	}
	if p.WindowSec <= 0 || math.IsInf(p.WindowSec, 0) || math.IsNaN(p.WindowSec) {
		return fmt.Errorf("window_sec must be > 0")
	}
	if p.TTLSec <= 0 {
		return fmt.Errorf("ttl_sec must be > 0")
	}
	return nil
}

// SlidingWindowAlgorithm implements the Redis Lua sliding-window-log contract.
type SlidingWindowAlgorithm struct{}

func (SlidingWindowAlgorithm) Name() string { return "sliding_window" }

func (SlidingWindowAlgorithm) KeyPrefix() string { return "ratelimit" }

func (SlidingWindowAlgorithm) BuildRedisKeys(baseKey string) []string {
	return []string{baseKey, baseKey + ":seq"}
}

func (SlidingWindowAlgorithm) UseRedisHashTag() bool { return true }

func (SlidingWindowAlgorithm) DefaultScriptPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("..", "..", "..", "scripts", "lua", "sliding_window.lua")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "scripts", "lua", "sliding_window.lua"))
}

func (SlidingWindowAlgorithm) ValidatePolicy(policy any) error {
	p, ok := asSlidingWindowPolicy(policy)
	if !ok {
		return fmt.Errorf("sliding_window algorithm requires SlidingWindowPolicy")
	}
	return p.Validate()
}

func (SlidingWindowAlgorithm) BuildRedisArgs(policy any, requested int64) ([]any, error) {
	p, ok := asSlidingWindowPolicy(policy)
	if !ok {
		return nil, fmt.Errorf("sliding_window algorithm requires SlidingWindowPolicy")
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	windowMs := int64(math.Ceil(p.WindowSec * 1000))
	return []any{p.Capacity, windowMs, requested, p.TTLSec}, nil
}

func (SlidingWindowAlgorithm) ParseDecision(raw any) (Decision, error) {
	return parseNormalizedDecision(raw)
}

func asSlidingWindowPolicy(policy any) (SlidingWindowPolicy, bool) {
	switch p := policy.(type) {
	case SlidingWindowPolicy:
		return p, true
	case *SlidingWindowPolicy:
		if p == nil {
			return SlidingWindowPolicy{}, false
		}
		return *p, true
	default:
		return SlidingWindowPolicy{}, false
	}
}
