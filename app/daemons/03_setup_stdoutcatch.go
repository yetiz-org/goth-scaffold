package daemons

import (
	kkdaemon "github.com/kklab-com/goth-daemon"
	kklogger "github.com/kklab-com/goth-kklogger"
	kkstdcatcher "github.com/kklab-com/goth-kkstdcatcher"
)

var DaemonSetupStdoutCatch = &SetupStdoutCatch{}

type SetupStdoutCatch struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupStdoutCatch) Start() {
	kkstdcatcher.StdoutWriteFunc = func(s string) {
		kklogger.InfoJ("STDOUT", s)
	}

	kkstdcatcher.StderrWriteFunc = func(s string) {
		kklogger.ErrorJ("STDERR", s)
	}

	kkstdcatcher.Start()
}
