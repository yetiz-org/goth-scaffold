// Task: default:log_current_time
// Goal: Log the current timestamp and write it as the task result; serves as the basic worker/queue health check template.
// Flow:
// 1. Get time.Now() and format as "2006-01-02 15:04:05".
// 2. kklogger.InfoJ with the formatted timestamp.
// 3. taskInfo.WriteSuccess with the same timestamp.

package tasks

import (
	"context"
	"time"

	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
	"github.com/yetiz-org/goth-scaffold/app/worker/tasks/taskdefs"
)

type LogCurrentTime struct {
	internal.DefaultHandler
}

func (l *LogCurrentTime) Name() string { return taskdefs.LogCurrentTime }

func (l *LogCurrentTime) Description() string {
	return "Log the current timestamp and write it as the task result; basic worker health check"
}

func (l *LogCurrentTime) Retention() time.Duration { return time.Hour }

func (l *LogCurrentTime) Run(ctx context.Context, taskInfo *internal.TaskInfo, executionContext map[string]any) error {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	kklogger.InfoJ("worker:LogCurrentTime.Run", currentTime)
	taskInfo.WriteSuccess(currentTime, nil)
	return nil
}
