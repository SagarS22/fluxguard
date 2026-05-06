package algorithms

import "testing"

func TestTokenBucketPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		policy  TokenBucketPolicy
		wantErr bool
	}{
		{"valid", TokenBucketPolicy{Capacity: 10, RefillRate: 1.5, TTLSec: 60}, false},
		{"bad capacity", TokenBucketPolicy{Capacity: 0, RefillRate: 1, TTLSec: 60}, true},
		{"bad refill", TokenBucketPolicy{Capacity: 10, RefillRate: -1, TTLSec: 60}, true},
		{"bad ttl", TokenBucketPolicy{Capacity: 10, RefillRate: 1, TTLSec: 0}, true},
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

func TestTokenBucketAlgorithmBuildRedisArgs(t *testing.T) {
	args, err := (TokenBucketAlgorithm{}).BuildRedisArgs(TokenBucketPolicy{Capacity: 10, RefillRate: 2.5, TTLSec: 30}, 3)
	if err != nil {
		t.Fatalf("BuildRedisArgs() error = %v", err)
	}
	want := []any{int64(10), 2.5, int64(3), int64(30)}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("arg[%d] = %#v, want %#v", i, args[i], want[i])
		}
	}
}

func TestTokenBucketAlgorithmParseDecision(t *testing.T) {
	decision, err := (TokenBucketAlgorithm{}).ParseDecision([]any{int64(1), int64(9), int64(0), int64(1234)})
	if err != nil {
		t.Fatalf("ParseDecision() error = %v", err)
	}
	if !decision.Allowed || decision.Remaining != 9 || decision.RetryAfterMs != 0 || decision.ResetAtMs != 1234 {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestTokenBucketAlgorithmParseDecisionRejectsBadResponse(t *testing.T) {
	_, err := (TokenBucketAlgorithm{}).ParseDecision([]any{1, 2, 3})
	if err == nil {
		t.Fatal("expected error")
	}
}
