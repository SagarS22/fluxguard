package ratelimiter

import (
	"fmt"
	"sync"

	"github.com/SagarS22/fluxguard/go/ratelimiter/algorithms"
)

// AlgorithmFactory constructs a new algorithm instance.
type AlgorithmFactory func() Algorithm

// AlgorithmRegistry stores named algorithm factories.
type AlgorithmRegistry struct {
	mu        sync.RWMutex
	factories map[string]AlgorithmFactory
}

// NewAlgorithmRegistry creates a registry with the built-in algorithms registered.
func NewAlgorithmRegistry() *AlgorithmRegistry {
	r := &AlgorithmRegistry{factories: make(map[string]AlgorithmFactory)}
	r.Register("token_bucket", func() Algorithm { return algorithms.TokenBucketAlgorithm{} })
	return r
}

// Register adds or replaces an algorithm factory.
func (r *AlgorithmRegistry) Register(name string, factory AlgorithmFactory) {
	if name == "" || factory == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Unregister removes an algorithm factory.
func (r *AlgorithmRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, name)
}

// Create constructs a named algorithm.
func (r *AlgorithmRegistry) Create(name string) (Algorithm, error) {
	r.mu.RLock()
	factory := r.factories[name]
	r.mu.RUnlock()
	if factory == nil {
		return nil, fmt.Errorf("%w: %s", ErrUnknownAlgorithm, name)
	}
	return factory(), nil
}

// Registry is the default process-wide algorithm registry.
var Registry = NewAlgorithmRegistry()
