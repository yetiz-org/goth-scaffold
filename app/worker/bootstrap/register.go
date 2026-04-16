package bootstrap

import (
	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/goth-scaffold/app/worker"
	"github.com/yetiz-org/goth-scaffold/app/worker/tasks"
	"github.com/yetiz-org/goth-scaffold/app/worker/tasks/taskdefs"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

// RegisterTasks registers all task handlers.
func RegisterTasks() {
	// Default handler
	worker.Register(&internal.DefaultHandler{})

	// Business tasks
	worker.Register(&tasks.LogCurrentTime{})
}

// RegisterScheduledTasks registers all cron tasks and returns their entry IDs.
func RegisterScheduledTasks(scheduler *asynq.Scheduler) []string {
	var entryIDs []string

	entryIDs = append(entryIDs, worker.RegisterSchedule(scheduler, "0 * * * *", worker.NewTask(taskdefs.LogCurrentTime, nil)))

	return entryIDs
}
