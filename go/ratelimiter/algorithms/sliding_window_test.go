package algorithms

import (
	"math"
	"os"
	"testing"
)

func TestSlidingWindowPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		policy  SlidingWindowPolicy
		wantErr bool
	}{
		{"valid", SlidingWindowPolicy{Capacity: 10, WindowSec: 60, TTLSec: 120}, false},
		{"bad capacity", SlidingWindowPolicy{Capacity: 0, WindowSec: 60, TTLSec: 120}, true},
		{"bad window zero", SlidingWindowPolicy{Capacity: 10, WindowSec: 0, TTLSec: 120}, true},
		{"bad window negative", SlidingWindowPolicy{Capacity: 10, WindowSec: -1, TTLSec: 120}, true},
		{"bad window infinity", SlidingWindowPolicy{Capacity: 10, WindowSec: math.Inf(1), TTLSec: 120}, true},
		{"bad window nan", SlidingWindowPolicy{Capacity: 10, WindowSec: math.NaN(), TTLSec: 120}, true},
		{"bad ttl", SlidingWindowPolicy{Capacity: 10, WindowSec: 60, TTLSec: 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlidingWindowAlgorithmBuildRedisArgs(t *testing.T) {
	args, err := (SlidingWindowAlgorithm{}).BuildRedisArgs(SlidingWindowPolicy{Capacity: 10, WindowSec: 1.001, TTLSec: 30}, 3)
	if err != nil {
		t.Fatalf("BuildRedisArgs() error = %v", err)
	}
	want := []any{int64(10), int64(1001), int64(3), int64(30)}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg[%d] = %#v, want %#v", i, args[i], want[i])
		}
	}
}

func TestSlidingWindowAlgorithmBuildRedisArgsCeilsSubsecondWindow(t *testing.T) {
	args, err := (SlidingWindowAlgorithm{}).BuildRedisArgs(SlidingWindowPolicy{Capacity: 10, WindowSec: 0.0001, TTLSec: 30}, 1)
	if err != nil {
		t.Fatalf("BuildRedisArgs() error = %v", err)
	}
	if got, want := args[1], int64(1); got != want {
		t.Fatalf("window_ms = %#v, want %#v", got, want)
	}
}

func TestSlidingWindowAlgorithmBuildRedisArgsRejectsWrongPolicy(t *testing.T) {
	_, err := (SlidingWindowAlgorithm{}).BuildRedisArgs(TokenBucketPolicy{Capacity: 10, RefillRate: 1, TTLSec: 30}, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSlidingWindowAlgorithmBuildRedisKeys(t *testing.T) {
	keys := (SlidingWindowAlgorithm{}).BuildRedisKeys("ratelimit:sliding_window:key:{dXNlcg}")
	want := []string{"ratelimit:sliding_window:key:{dXNlcg}", "ratelimit:sliding_window:key:{dXNlcg}:seq"}
	if !equalStringSlices(keys, want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
}

func TestSlidingWindowAlgorithmParseDecision(t *testing.T) {
	decision, err := (SlidingWindowAlgorithm{}).ParseDecision([]any{int64(1), int64(9), int64(0), int64(1234)})
	if err != nil {
		t.Fatalf("ParseDecision() error = %v", err)
	}
	if !decision.Allowed || decision.Remaining != 9 || decision.RetryAfterMs != 0 || decision.ResetAtMs != 1234 {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestSlidingWindowAlgorithmParseDecisionRejectsBadResponse(t *testing.T) {
	_, err := (SlidingWindowAlgorithm{}).ParseDecision([]any{1, 2, 3})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSlidingWindowAlgorithmDefaultScriptPathExists(t *testing.T) {
	path := (SlidingWindowAlgorithm{}).DefaultScriptPath()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("DefaultScriptPath() = %q, stat error = %v", path, err)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
