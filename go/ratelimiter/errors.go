package ratelimiter

import "errors"

var (
	ErrScriptExecution  = errors.New("script execution failed")
	ErrUnknownAlgorithm = errors.New("unknown algorithm")
	ErrInvalidPolicy    = errors.New("invalid policy")
	ErrInvalidResponse  = errors.New("invalid script response")
)
