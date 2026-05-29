package script

import (
	"context"
	"errors"
	"testing"
)

type fakeRedis struct {
	loads      int
	evals      int
	loadErr    error
	evalErrs   []error
	evalResult any
	evalKeys   [][]string
}

func (f *fakeRedis) ScriptLoad(context.Context, string) (string, error) {
	f.loads++
	if f.loadErr != nil {
		return "", f.loadErr
	}
	return "sha", nil
}

func (f *fakeRedis) EvalSHA(_ context.Context, _ string, keys []string, _ ...any) (any, error) {
	f.evals++
	f.evalKeys = append(f.evalKeys, append([]string(nil), keys...))
	if len(f.evalErrs) > 0 {
		err := f.evalErrs[0]
		f.evalErrs = f.evalErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	return f.evalResult, nil
}

func TestExecutorLoadsAndExecutes(t *testing.T) {
	redis := &fakeRedis{evalResult: []any{1, 2, 3, 4}}
	executor, err := NewExecutor(redis, "", "return {1,2,3,4}")
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	_, err = executor.Execute(context.Background(), "key", []any{1})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if redis.loads != 1 || redis.evals != 1 {
		t.Fatalf("loads/evals = %d/%d, want 1/1", redis.loads, redis.evals)
	}
	if got, want := redis.evalKeys[0], []string{"key"}; !equalStrings(got, want) {
		t.Fatalf("eval keys = %v, want %v", got, want)
	}
}

func TestExecutorPassesMultipleKeys(t *testing.T) {
	redis := &fakeRedis{evalResult: []any{1, 2, 3, 4}}
	executor, err := NewExecutor(redis, "", "return {1,2,3,4}")
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	_, err = executor.ExecuteKeys(context.Background(), []string{"log-key", "seq-key"}, []any{1})
	if err != nil {
		t.Fatalf("ExecuteKeys() error = %v", err)
	}
	if redis.loads != 1 || redis.evals != 1 {
		t.Fatalf("loads/evals = %d/%d, want 1/1", redis.loads, redis.evals)
	}
	if got, want := redis.evalKeys[0], []string{"log-key", "seq-key"}; !equalStrings(got, want) {
		t.Fatalf("eval keys = %v, want %v", got, want)
	}
}

func TestExecutorReloadsOnNoScript(t *testing.T) {
	redis := &fakeRedis{evalErrs: []error{errors.New("NOSCRIPT no matching script")}, evalResult: []any{1, 2, 3, 4}}
	executor, err := NewExecutor(redis, "", "return {1,2,3,4}")
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	keys := []string{"log-key", "seq-key"}
	_, err = executor.ExecuteKeys(context.Background(), keys, nil)
	if err != nil {
		t.Fatalf("ExecuteKeys() error = %v", err)
	}
	if redis.loads != 2 || redis.evals != 2 {
		t.Fatalf("loads/evals = %d/%d, want 2/2", redis.loads, redis.evals)
	}
	if len(redis.evalKeys) != 2 {
		t.Fatalf("eval key calls = %d, want 2", len(redis.evalKeys))
	}
	for i, got := range redis.evalKeys {
		if !equalStrings(got, keys) {
			t.Fatalf("eval keys call %d = %v, want %v", i, got, keys)
		}
	}
}

func TestExecutorRejectsEmptyScript(t *testing.T) {
	_, err := NewExecutor(&fakeRedis{}, "", "   ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
