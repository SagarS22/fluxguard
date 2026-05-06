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
}

func (f *fakeRedis) ScriptLoad(context.Context, string) (string, error) {
	f.loads++
	if f.loadErr != nil {
		return "", f.loadErr
	}
	return "sha", nil
}

func (f *fakeRedis) EvalSHA(context.Context, string, []string, ...any) (any, error) {
	f.evals++
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
}

func TestExecutorReloadsOnNoScript(t *testing.T) {
	redis := &fakeRedis{evalErrs: []error{errors.New("NOSCRIPT no matching script")}, evalResult: []any{1, 2, 3, 4}}
	executor, err := NewExecutor(redis, "", "return {1,2,3,4}")
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	_, err = executor.Execute(context.Background(), "key", nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if redis.loads != 2 || redis.evals != 2 {
		t.Fatalf("loads/evals = %d/%d, want 2/2", redis.loads, redis.evals)
	}
}

func TestExecutorRejectsEmptyScript(t *testing.T) {
	_, err := NewExecutor(&fakeRedis{}, "", "   ")
	if err == nil {
		t.Fatal("expected error")
	}
}
