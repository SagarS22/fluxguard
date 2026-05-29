package algorithms

import (
	"fmt"
	"math"
	"strconv"
)

func parseNormalizedDecision(raw any) (Decision, error) {
	parts, ok := toSlice(raw)
	if !ok || len(parts) != 4 {
		return Decision{}, fmt.Errorf("unexpected script response: %v", raw)
	}

	allowed, err := toInt64(parts[0])
	if err != nil {
		return Decision{}, fmt.Errorf("invalid allowed value: %w", err)
	}
	remaining, err := toInt64(parts[1])
	if err != nil {
		return Decision{}, fmt.Errorf("invalid remaining value: %w", err)
	}
	retryAfter, err := toInt64(parts[2])
	if err != nil {
		return Decision{}, fmt.Errorf("invalid retry_after_ms value: %w", err)
	}
	resetAt, err := toInt64(parts[3])
	if err != nil {
		return Decision{}, fmt.Errorf("invalid reset_at_ms value: %w", err)
	}

	return Decision{
		Allowed:      allowed != 0,
		Remaining:    remaining,
		RetryAfterMs: retryAfter,
		ResetAtMs:    resetAt,
	}, nil
}

func toSlice(raw any) ([]any, bool) {
	switch v := raw.(type) {
	case []any:
		return v, true
	case []int:
		out := make([]any, len(v))
		for i, n := range v {
			out[i] = n
		}
		return out, true
	case []int64:
		out := make([]any, len(v))
		for i, n := range v {
			out[i] = n
		}
		return out, true
	default:
		return nil, false
	}
}

func toInt64(v any) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		if uint64(n) > math.MaxInt64 {
			return 0, fmt.Errorf("uint overflows int64")
		}
		return int64(n), nil
	case uint8:
		return int64(n), nil
	case uint16:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		if n > math.MaxInt64 {
			return 0, fmt.Errorf("uint64 overflows int64")
		}
		return int64(n), nil
	case float64:
		if math.Trunc(n) != n {
			return 0, fmt.Errorf("float64 is not integral")
		}
		return int64(n), nil
	case string:
		return strconv.ParseInt(n, 10, 64)
	case []byte:
		return strconv.ParseInt(string(n), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported integer type %T", v)
	}
}
