package internal

import (
	"context"
	"encoding/json"
	"io"

	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/gone/ghttp"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

func NewTask(typename string, payload any, opts ...asynq.Option) *asynq.Task {
	var encoded []byte
	if payload != nil {
		var err error
		encoded, err = json.Marshal(payload)
		if err != nil {
			kklogger.ErrorJ("internal:NewTask#marshal!failed", err.Error())
		}
	}

	return asynq.NewTask(typename, encoded, opts...)
}

// TaskInfo task information
type TaskInfo struct {
	Name          string `json:"n"`
	Payload       any    `json:"p"`
	_ResultWriter io.Writer
}

// Write implements io.Writer interface for Task result output
func (t *TaskInfo) Write(data []byte) (n int, err error) {
	if t._ResultWriter == nil {
		return 0, nil
	}
	return t._ResultWriter.Write(data)
}

func NewTaskInfo(name string, payload any, resultWriter io.Writer) *TaskInfo {
	return &TaskInfo{
		Name:          name,
		Payload:       payload,
		_ResultWriter: resultWriter,
	}
}

// Handler task handler interface
type Handler interface {
	Name() string
	Before(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	Run(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	After(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any) error
	ErrorCaught(ctx context.Context, taskInfo *TaskInfo, executionContext map[string]any, err ghttp.ErrorResponse) error
}

// DefaultHandler default handler
type DefaultHandler struct{}

func (j *DefaultHandler) Name() string {
	return "default"
}

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
