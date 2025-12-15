package worker

import (
	"encoding/json"
	"fmt"

	"github.com/yetiz-org/asynq"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

func NewTask(typename string, payload any, opts ...asynq.Option) *asynq.Task {
	var encoded []byte
	if payload != nil {
		var err error
		encoded, err = json.Marshal(payload)
		if err != nil {
			kklogger.ErrorJ("worker:NewTask#marshal!failed", err.Error())
		}
	}

	return asynq.NewTask(typename, encoded, opts...)
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
