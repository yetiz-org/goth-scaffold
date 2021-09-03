package daemons

import (
	"time"
)

var DaemonLoopExample = WrapTimerDaemon(new(LoopExample))

type LoopExample struct {
	DefaultTimerDaemon
}

func (d *LoopExample) Interval() time.Duration {
	return time.Minute
}

func (d *LoopExample) Loop() error {
	// do something
	return nil
}
