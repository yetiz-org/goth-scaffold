package daemons

import (
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
)

var DaemonSchedulerLoopExample = new(SchedulerLoopExample)

type SchedulerLoopExample struct {
	kkdaemon.DefaultSchedulerDaemon
}

func (d *SchedulerLoopExample) Registered() error {
	// init func
	return nil
}

func (d *SchedulerLoopExample) Start() {
	// do when start
}

func (d *SchedulerLoopExample) When() kkdaemon.CronSyntax {
	// run every two minute
	return "*/2 * * * *"
}

func (d *SchedulerLoopExample) Loop() error {
	// do something
	return nil
}

func (d *SchedulerLoopExample) Stop(sig os.Signal) {
	// do when stop
}
