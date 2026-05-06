package algorithms

// Decision is the normalized result returned by rate-limit algorithms.
type Decision struct {
	Allowed      bool
	Remaining    int64
	RetryAfterMs int64
	ResetAtMs    int64
}

// Algorithm is the extension contract used by the root RateLimiter.
type Algorithm interface {
	Name() string
	KeyPrefix() string
	DefaultScriptPath() string
	ValidatePolicy(policy any) error
	BuildRedisArgs(policy any, requested int64) ([]any, error)
	ParseDecision(raw any) (Decision, error)
}
