// Package ratelimiter provides an extensible Redis/Lua-backed rate limiter SDK.
//
// The default algorithm is token_bucket. Custom algorithms can be provided via
// WithAlgorithm or registered in Registry and constructed by name with
// CreateRateLimiter.
package ratelimiter
