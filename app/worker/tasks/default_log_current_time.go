package tasks

import (
	"context"
	"time"

	"github.com/yetiz-org/asynq"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

// LogCurrentTime logs current time
type LogCurrentTime struct {
	internal.DefaultHandler
}

func (l *LogCurrentTime) Name() string {
	return "default:log_current_time"
}

// NewLogCurrentTimeTask creates a task that logs current time
// Example:
//
//	task := tasks.NewLogCurrentTimeTask()
//	worker.Submit(task)
func NewLogCurrentTimeTask() *asynq.Task {
	return internal.NewTask("default:log_current_time", nil, asynq.Retention(60*time.Second))
}

// Run logs current time
func (l *LogCurrentTime) Run(ctx context.Context, taskInfo *internal.TaskInfo, executionContext map[string]any) error {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	kklogger.InfoJ("worker:LogCurrentTime.Run", currentTime)
	taskInfo.Write([]byte(currentTime))
	return nil
}
