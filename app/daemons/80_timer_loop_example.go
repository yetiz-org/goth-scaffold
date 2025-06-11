package daemons

import (
	"os"
	"time"

	kkdaemon "github.com/yetiz-org/goth-daemon"
)

var DaemonTimerLoopExample = new(TimerLoopExample)

type TimerLoopExample struct {
	kkdaemon.DefaultTimerDaemon
}

func (d *TimerLoopExample) Registered() error {
	// init func
	return nil
}

func (d *TimerLoopExample) Start() {
	// do when start
}

func (d *TimerLoopExample) Interval() time.Duration {
	// run every duration
	return time.Minute
}

func (d *TimerLoopExample) Loop() error {
	// do something
	return nil
}

func (d *TimerLoopExample) Stop(sig os.Signal) {
	// do when stop
}
