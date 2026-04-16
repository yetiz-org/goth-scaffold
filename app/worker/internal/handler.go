package internal

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
	"time"

	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/gone/ghttp"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

type Payload interface {
	Encode() (string, error)
	Decode(data string) error
	Validate() error
}

// DryRunSupporter marks a payload as supporting dry_run.
// Payloads implementing this interface will expose the BasePayload dry_run field in API output.
type DryRunSupporter interface {
	SupportDryRun()
}

// BasePayload implements the Payload interface. Embed it to get default behaviour.
// Override Encode/Decode when the struct has additional fields; call EncodePayload/DecodePayload.
type BasePayload struct {
	DryRun bool `json:"dry_run,omitempty" required:"false" desc:"If true, only outputs results without any database writes or external side effects"`
}

func (p *BasePayload) Encode() (string, error) { return "{}", nil }
func (p *BasePayload) Decode(string) error     { return nil }
func (p *BasePayload) Validate() error         { return nil }

func (p *BasePayload) EncodePayload(self any) (string, error) {
	data, err := json.Marshal(self)
	return string(data), err
}

func (p *BasePayload) DecodePayload(data string, self any) error {
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal([]byte(data), self)
}

func NewTask(taskname string, payload Payload) *asynq.Task {
	var encoded []byte
	if payload != nil && !reflect.ValueOf(payload).IsNil() {
		str, err := payload.Encode()
		if err != nil {
			kklogger.ErrorJ("internal:NewTask#encode!failed", err.Error())
		}

		encoded = []byte(str)
	}

	return asynq.NewTask(taskname, encoded)
}

// taskResultEnvelope is the internal wrapper; WriteSuccess/WriteError auto-populate timing fields.
type taskResultEnvelope struct {
	Success     bool   `json:"success"`
	DurationMs  int64  `json:"duration_ms"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
	Data        any    `json:"data,omitempty"`
	StartedAt   int64  `json:"started_at"`
	CompletedAt int64  `json:"completed_at"`
}

// pendingResult is the staged result, flushed when FlushResult is called.
type pendingResult struct {
	success bool
	errCode string
	message string
	data    any
}

// TaskInfo carries task metadata through the handler lifecycle.
type TaskInfo struct {
	Name           string  `json:"n"`
	Payload        Payload `json:"p"`
	_ResultWriter  io.Writer
	_startedAt     time.Time
	_pendingResult *pendingResult
}

// Write implements io.Writer for task result output.
func (t *TaskInfo) Write(data []byte) (n int, err error) {
	if t._ResultWriter == nil {
		return 0, nil
	}
	return t._ResultWriter.Write(data)
}

func NewTaskInfo(name string, payload Payload, resultWriter io.Writer) *TaskInfo {
	return &TaskInfo{
		Name:          name,
		Payload:       payload,
		_ResultWriter: resultWriter,
		_startedAt:    time.Now(),
	}
}

// WriteSuccess stages a success result; written on FlushResult.
func (t *TaskInfo) WriteSuccess(message string, data any) {
	t._pendingResult = &pendingResult{
		success: true,
		message: message,
		data:    data,
	}
}

func (t *TaskInfo) WriteDryRunSuccess(data any) {
	t.WriteSuccess("dry_run", data)
}

// WriteError stages a failure result; written on FlushResult.
func (t *TaskInfo) WriteError(errCode string, message string, data any) {
	t._pendingResult = &pendingResult{
		success: false,
		errCode: errCode,
		message: message,
		data:    data,
	}
}

// FlushResult is called at the end of ProcessTask; computes final timing and writes the result.
func (t *TaskInfo) FlushResult() error {
	if t._pendingResult == nil {
		return nil
	}

	now := time.Now()
	envelope := &taskResultEnvelope{
		Success:     t._pendingResult.success,
		Error:       t._pendingResult.errCode,
		Message:     t._pendingResult.message,
		Data:        t._pendingResult.data,
		StartedAt:   t._startedAt.Unix(),
		CompletedAt: now.Unix(),
		DurationMs:  now.Sub(t._startedAt).Milliseconds(),
	}

	jsonData, err := json.Marshal(envelope)
	if err != nil {
		kklogger.ErrorJ("internal:TaskInfo.FlushResult#marshal!failed", err.Error())
		return err
	}

	_, err = t.Write(jsonData)
	t._pendingResult = nil
	return err
}

// Handler is the task handler interface.
type Handler interface {
	Name() string
	Description() string      // One-line task description
	Retention() time.Duration // How long to keep the result; 0 = discard
	UniqueTTL() time.Duration // Unique dedup window; 0 = disabled
	TaskID() string           // Fixed task ID; "" = system-generated
	Timeout() time.Duration   // Execution timeout; 0 = asynq server default
	Queue() string            // Queue name; "" = default queue
	MaxRetry() int            // Max retries; -1 = asynq default (25); 0 = no retry
	Group() string            // Task group; "" = no group
	Payload() Payload         // Returns an empty Payload instance for this handler
	Before(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	Run(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	After(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	ErrorCaught(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any, err ghttp.ErrorResponse) error
}

// DefaultHandler provides sensible defaults for all Handler methods.
type DefaultHandler struct{}

func (j *DefaultHandler) Name() string { return "default" }

func (j *DefaultHandler) Description() string { return "" }

func (j *DefaultHandler) Retention() time.Duration { return 0 }

func (j *DefaultHandler) UniqueTTL() time.Duration { return 0 }

func (j *DefaultHandler) TaskID() string { return "" }

func (j *DefaultHandler) Timeout() time.Duration { return 60 * time.Minute }

func (j *DefaultHandler) Queue() string { return "" }

func (j *DefaultHandler) MaxRetry() int { return 3 }

func (j *DefaultHandler) Group() string { return "" }

func (j *DefaultHandler) Payload() Payload { return &BasePayload{} }

func (j *DefaultHandler) Before(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error {
	return nil
}

func (j *DefaultHandler) Run(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error {
	return nil
}

func (j *DefaultHandler) After(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error {
	return nil
}

func (j *DefaultHandler) ErrorCaught(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any, err ghttp.ErrorResponse) error {
	m := map[string]any{}
	if jErr := json.Unmarshal([]byte(err.Message()), &m); jErr == nil {
		err.ErrorData()[".message"] = m
		err.ErrorData()[".task_info"] = taskInfo
	}
	return err
}
