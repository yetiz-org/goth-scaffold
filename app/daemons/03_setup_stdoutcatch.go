package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	kkstdcatcher "github.com/yetiz-org/goth-kkstdcatcher"
)

var DaemonSetupStdoutCatch = &SetupStdoutCatch{}

type SetupStdoutCatch struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupStdoutCatch) Start() {
	kkstdcatcher.DefaultInstance().StdoutWriteFunc = func(s string) {
		kklogger.InfoJ("STDOUT", s)
	}

	kkstdcatcher.DefaultInstance().StderrWriteFunc = func(s string) {
		kklogger.ErrorJ("STDERR", s)
	}

	kkstdcatcher.Start()
}
