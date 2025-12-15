package worker

import (
	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
	"github.com/yetiz-org/goth-scaffold/app/worker/tasks"
)

// _RegisterTasks registers all task handlers
func _RegisterTasks() {
	// Default tasks
	Register("", &internal.DefaultHandler{})

	// Business tasks
	Register((&tasks.LogCurrentTime{}).Name(), &tasks.LogCurrentTime{})
}

// _RegisterScheduledTasks registers all scheduled tasks and returns entry IDs
func _RegisterScheduledTasks(scheduler *asynq.Scheduler) []string {
	var entryIDs []string

	entryIDs = append(entryIDs, RegisterSchedule(scheduler, "0 * * * *", tasks.NewLogCurrentTimeTask()))
	return entryIDs
}
