package ratelimiter

import (
	"strings"
	"testing"
)

func TestEncodeKeyComponent(t *testing.T) {
	encoded, err := encodeKeyComponent("user:123")
	if err != nil {
		t.Fatalf("encodeKeyComponent() error = %v", err)
	}
	if encoded != "dXNlcjoxMjM" {
		t.Fatalf("encoded = %q, want %q", encoded, "dXNlcjoxMjM")
	}
	if strings.Contains(encoded, "=") {
		t.Fatalf("encoded key contains padding: %q", encoded)
	}
}

func TestEncodeKeyComponentUsesURLSafeAlphabet(t *testing.T) {
	encoded, err := encodeKeyComponent(string([]byte{0xfb, 0xff}))
	if err != nil {
		t.Fatalf("encodeKeyComponent() error = %v", err)
	}
	if encoded != "-_8" {
		t.Fatalf("encoded = %q, want URL-safe %q", encoded, "-_8")
	}
	if strings.ContainsAny(encoded, "+/") {
		t.Fatalf("encoded key uses non-URL-safe alphabet: %q", encoded)
	}
}

func TestRedisKeyBuilderBuildBaseKey(t *testing.T) {
	builder := newRedisKeyBuilder()
	key, err := builder.buildBaseKey("ratelimit", "token_bucket", "key", "user:123", false)
	if err != nil {
		t.Fatalf("buildBaseKey() error = %v", err)
	}
	if key != "ratelimit:token_bucket:key:dXNlcjoxMjM" {
		t.Fatalf("key = %q", key)
	}
}

func TestRedisKeyBuilderDefaultsDimension(t *testing.T) {
	builder := newRedisKeyBuilder()
	key, err := builder.buildBaseKey("ratelimit", "token_bucket", "", "user:123", false)
	if err != nil {
		t.Fatalf("buildBaseKey() error = %v", err)
	}
	if key != "ratelimit:token_bucket:key:dXNlcjoxMjM" {
		t.Fatalf("key = %q", key)
	}
}

func TestRedisKeyBuilderWrapsHashTag(t *testing.T) {
	builder := newRedisKeyBuilder()
	key, err := builder.buildBaseKey("ratelimit", "sliding_window", "key", "user:123", true)
	if err != nil {
		t.Fatalf("buildBaseKey() error = %v", err)
	}
	if key != "ratelimit:sliding_window:key:{dXNlcjoxMjM}" {
		t.Fatalf("key = %q", key)
	}
}

func TestRedisKeyBuilderRejectsInvalidStructuralComponents(t *testing.T) {
	builder := newRedisKeyBuilder()
	tests := []struct {
		name      string
		prefix    string
		algorithm string
		dimension string
	}{
		{"bad prefix", "rate:limit", "token_bucket", "key"},
		{"bad algorithm", "ratelimit", "token.bucket", "key"},
		{"bad dimension", "ratelimit", "token_bucket", "user:id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := builder.buildBaseKey(tt.prefix, tt.algorithm, tt.dimension, "user", false)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRedisKeyBuilderTrimsStructuralComponentColons(t *testing.T) {
	builder := newRedisKeyBuilder()
	key, err := builder.buildBaseKey(":ratelimit:", ":token_bucket:", ":key:", "user:123", false)
	if err != nil {
		t.Fatalf("buildBaseKey() error = %v", err)
	}
	if key != "ratelimit:token_bucket:key:dXNlcjoxMjM" {
		t.Fatalf("key = %q", key)
	}
}

func TestRedisKeyBuilderRejectsOverMaxLength(t *testing.T) {
	builder := newRedisKeyBuilder()
	_, err := builder.buildBaseKey("ratelimit", "token_bucket", "key", strings.Repeat("x", 400), false)
	if err == nil {
		t.Fatal("expected error")
	}
}
