package ratelimiter

import "strings"

type config struct {
	algorithm    Algorithm
	keyPrefix    string
	scriptPath   string
	scriptSource string
}

// Option customizes RateLimiter construction.
type Option func(*config) error

// WithAlgorithm selects an algorithm implementation.
func WithAlgorithm(algorithm Algorithm) Option {
	return func(c *config) error {
		if algorithm == nil {
			return ErrUnknownAlgorithm
		}
		c.algorithm = algorithm
		return nil
	}
}

// WithKeyPrefix overrides the algorithm default Redis key prefix.
func WithKeyPrefix(prefix string) Option {
	return func(c *config) error {
		c.keyPrefix = strings.Trim(prefix, ":")
		return nil
	}
}

// WithScriptPath overrides the algorithm default Lua script path.
func WithScriptPath(path string) Option {
	return func(c *config) error {
		c.scriptPath = path
		return nil
	}
}

// WithScriptSource supplies Lua script source directly, useful in tests.
func WithScriptSource(source string) Option {
	return func(c *config) error {
		c.scriptSource = source
		return nil
	}
}
