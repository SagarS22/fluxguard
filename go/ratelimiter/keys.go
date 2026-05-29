package ratelimiter

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

const (
	defaultRedisKeyDimension = "key"
	defaultMaxRedisKeyBytes  = 512
)

var structuralComponentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// redisKeyBuilder builds ADR-compliant Redis keys for rate-limiter scripts.
type redisKeyBuilder struct {
	maxKeyBytes int
}

func newRedisKeyBuilder() redisKeyBuilder {
	return redisKeyBuilder{maxKeyBytes: defaultMaxRedisKeyBytes}
}

func encodeKeyComponent(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("key must be non-empty")
	}
	return base64.RawURLEncoding.EncodeToString([]byte(value)), nil
}

func (b redisKeyBuilder) buildBaseKey(prefix, algorithm, dimension, key string, hashTag bool) (string, error) {
	prefix, err := validateStructuralComponent("prefix", prefix)
	if err != nil {
		return "", err
	}
	algorithm, err = validateStructuralComponent("algorithm", algorithm)
	if err != nil {
		return "", err
	}
	if dimension == "" {
		dimension = defaultRedisKeyDimension
	}
	dimension, err = validateStructuralComponent("dimension", dimension)
	if err != nil {
		return "", err
	}

	encodedKey, err := encodeKeyComponent(key)
	if err != nil {
		return "", err
	}
	keyComponent := encodedKey
	if hashTag {
		keyComponent = "{" + encodedKey + "}"
	}

	redisKey := prefix + ":" + algorithm + ":" + dimension + ":" + keyComponent
	if err := b.validateKeyLength(redisKey); err != nil {
		return "", err
	}
	return redisKey, nil
}

func (b redisKeyBuilder) validateKeyLength(redisKey string) error {
	maxKeyBytes := b.maxKeyBytes
	if maxKeyBytes == 0 {
		maxKeyBytes = defaultMaxRedisKeyBytes
	}
	if len([]byte(redisKey)) > maxKeyBytes {
		return fmt.Errorf("redis key exceeds %d bytes", maxKeyBytes)
	}
	return nil
}

func validateStructuralComponent(name, value string) (string, error) {
	value = strings.Trim(value, ":")
	if value == "" {
		return "", fmt.Errorf("%s must be non-empty", name)
	}
	if !structuralComponentPattern.MatchString(value) {
		return "", fmt.Errorf("%s may only contain letters, numbers, '_' and '-'", name)
	}
	return value, nil
}
