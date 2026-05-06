package ratelimiter

import "github.com/SagarS22/fluxguard/go/ratelimiter/algorithms"

// Decision is the normalized rate-limit decision returned by Check.
type Decision = algorithms.Decision

// Algorithm is the extension contract implemented by rate-limit algorithms.
type Algorithm = algorithms.Algorithm

// TokenBucketPolicy configures the built-in token-bucket algorithm.
type TokenBucketPolicy = algorithms.TokenBucketPolicy

// TokenBucketAlgorithm is the built-in token-bucket algorithm implementation.
type TokenBucketAlgorithm = algorithms.TokenBucketAlgorithm
