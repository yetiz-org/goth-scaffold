package worker

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

// _execSuccessHandler stages a success result and returns nil.
type _execSuccessHandler struct {
	internal.DefaultHandler
}

func (h *_execSuccessHandler) Name() string { return "test:execute_success" }

func (h *_execSuccessHandler) Run(_ context.Context, taskInfo *internal.TaskInfo, _ map[string]any) error {
	taskInfo.WriteSuccess("ok", nil)
	return nil
}

// _execErrorHandler stages an error result and returns a non-nil error.
type _execErrorHandler struct {
	internal.DefaultHandler
}

func (h *_execErrorHandler) Name() string { return "test:execute_error" }

func (h *_execErrorHandler) Run(_ context.Context, taskInfo *internal.TaskInfo, _ map[string]any) error {
	taskInfo.WriteError("E_BOOM", "boom", nil)
	return errors.New("boom")
}

// Execute must return the structured result on the synchronous path, where there is no Redis
// ResultWriter (nil writer). Regression guard: Result() must cache the envelope before
// FlushResult clears the staged result, otherwise the post-flush read returns nil.
func TestExecute_SuccessReturnsStructuredResultWithNilResultWriter(t *testing.T) {
	Register(&_execSuccessHandler{})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Execute() panicked with nil ResultWriter: %v", r)
		}
	}()

	result, err := Execute(context.Background(), NewTask("test:execute_success", nil))
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("Execute() result = nil, want structured result even with nil ResultWriter")
	}

	if !result.Success {
		t.Errorf("result.Success = false, want true")
	}

	if result.Message != "ok" {
		t.Errorf("result.Message = %q, want %q", result.Message, "ok")
	}
}

// An unregistered task type yields asynq.NotFound and no result.
func TestExecute_UnregisteredTaskTypeReturnsNotFound(t *testing.T) {
	result, err := Execute(context.Background(), NewTask("test:execute_unregistered", nil))
	if result != nil {
		t.Fatalf("Execute() result = %+v, want nil for unregistered task", result)
	}

	if err == nil {
		t.Fatal("Execute() error = nil, want handler-not-found error")
	}

	if !strings.Contains(err.Error(), "handler not found") {
		t.Errorf("Execute() error = %q, want it to mention 'handler not found'", err.Error())
	}
}

// On the synchronous path a failing handler still returns its staged error result alongside the
// error. Here the result is produced solely by Result() — FlushResult never runs in the error path.
func TestExecute_HandlerErrorReturnsStructuredErrorResult(t *testing.T) {
	Register(&_execErrorHandler{})

	result, err := Execute(context.Background(), NewTask("test:execute_error", nil))
	if err == nil {
		t.Fatal("Execute() error = nil, want handler error")
	}

	if result == nil {
		t.Fatal("Execute() result = nil, want staged error result")
	}

	if result.Success {
		t.Errorf("result.Success = true, want false")
	}

	if result.Error != "E_BOOM" {
		t.Errorf("result.Error = %q, want %q", result.Error, "E_BOOM")
	}
}
