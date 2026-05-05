package ratelimiter

// CreateRateLimiter constructs a RateLimiter using an algorithm registered by name.
func CreateRateLimiter(client RedisClient, algorithmName string, opts ...Option) (*RateLimiter, error) {
	algorithm, err := Registry.Create(algorithmName)
	if err != nil {
		return nil, err
	}
	return New(client, append([]Option{WithAlgorithm(algorithm)}, opts...)...)
}
