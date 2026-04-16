package worker

import (
	"fmt"

	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

// GetPayloadByTaskname returns an empty Payload instance for the given task name.
// Used by external callers that need to decode a task's payload.
func GetPayloadByTaskname(taskname string) internal.Payload {
	if handler, ok := handlerRegistry[taskname]; ok {
		return handler.Payload()
	}

	return &internal.BasePayload{}
}

// NewTask creates an asynq.Task from a task name and Payload.
// Pass nil payload when no payload is needed.
func NewTask(taskname string, payload internal.Payload) *asynq.Task {
	return internal.NewTask(taskname, payload)
}

type TaskFuture struct {
	inspector *asynq.Inspector
	*asynq.TaskInfo
}

func (t *TaskFuture) ID() string {
	return fmt.Sprintf("%s:%s", t.TaskInfo.Queue, t.TaskInfo.ID)
}

func (t *TaskFuture) Update() error {
	if info, err := t.inspector.GetTaskInfo(t.TaskInfo.Queue, t.TaskInfo.ID); err != nil {
		return err
	} else {
		t.TaskInfo = info
	}

	return nil
}

func NewTaskFuture(inspector *asynq.Inspector, taskInfo *asynq.TaskInfo) *TaskFuture {
	return &TaskFuture{
		inspector: inspector,
		TaskInfo:  taskInfo,
	}
}
