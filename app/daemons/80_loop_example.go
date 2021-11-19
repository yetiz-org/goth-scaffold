package daemons

import (
	"os"
	"time"

	kkdaemon "github.com/kklab-com/goth-daemon"
)

var DaemonLoopExample = new(LoopExample)

type LoopExample struct {
	kkdaemon.DefaultTimerDaemon
}

func (d *LoopExample) Registered() error {
	// init func
	return nil
}

func (d *LoopExample) Start() {
	// do when start
}

func (d *LoopExample) Interval() time.Duration {
	// run every duration
	return time.Minute
}

func (d *LoopExample) Loop() error {
	// do something
	return nil
}

func (d *LoopExample) Stop(sig os.Signal) {
	// do when stop
}
