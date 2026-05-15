package script

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// RedisClient is the minimal Redis scripting API required by the SDK.
type RedisClient interface {
	ScriptLoad(ctx context.Context, script string) (string, error)
	EvalSHA(ctx context.Context, sha string, keys []string, args ...any) (any, error)
}

// Executor owns Lua script source loading plus SCRIPT LOAD/EVALSHA lifecycle.
type Executor struct {
	redis  RedisClient
	source string

	mu  sync.Mutex
	sha string
}

// NewExecutor creates an executor from either scriptPath or scriptSource.
func NewExecutor(redis RedisClient, scriptPath, scriptSource string) (*Executor, error) {
	if redis == nil {
		return nil, fmt.Errorf("redis client must not be nil")
	}

	source, err := loadScriptSource(scriptPath, scriptSource)
	if err != nil {
		return nil, err
	}

	return &Executor{redis: redis, source: source}, nil
}

func loadScriptSource(scriptPath, scriptSource string) (string, error) {
	var source string
	switch {
	case scriptSource != "":
		source = scriptSource
	case scriptPath != "":
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			return "", fmt.Errorf("read lua script: %w", err)
		}
		source = string(content)
	default:
		return "", fmt.Errorf("script_path or script_source must be provided")
	}

	if strings.TrimSpace(source) == "" {
		return "", fmt.Errorf("lua script is empty")
	}
	return source, nil
}

// Execute runs the loaded script against one Redis key.
func (e *Executor) Execute(ctx context.Context, redisKey string, args []any) (any, error) {
	return e.ExecuteKeys(ctx, []string{redisKey}, args)
}

// ExecuteKeys runs the loaded script against one or more Redis keys.
func (e *Executor) ExecuteKeys(ctx context.Context, redisKeys []string, args []any) (any, error) {
	sha, err := e.ensureScriptLoaded(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := e.redis.EvalSHA(ctx, sha, redisKeys, args...)
	if err == nil {
		return raw, nil
	}
	if !isNoScript(err) {
		return nil, fmt.Errorf("evalsha: %w", err)
	}

	e.mu.Lock()
	e.sha = ""
	e.mu.Unlock()

	sha, err = e.ensureScriptLoaded(ctx)
	if err != nil {
		return nil, fmt.Errorf("reload after NOSCRIPT: %w", err)
	}
	raw, err = e.redis.EvalSHA(ctx, sha, redisKeys, args...)
	if err != nil {
		return nil, fmt.Errorf("evalsha after NOSCRIPT reload: %w", err)
	}
	return raw, nil
}

func (e *Executor) ensureScriptLoaded(ctx context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.sha != "" {
		return e.sha, nil
	}
	sha, err := e.redis.ScriptLoad(ctx, e.source)
	if err != nil {
		return "", fmt.Errorf("script load: %w", err)
	}
	e.sha = sha
	return sha, nil
}

func isNoScript(err error) bool {
	return strings.Contains(strings.ToUpper(err.Error()), "NOSCRIPT")
}
