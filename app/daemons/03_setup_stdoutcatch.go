package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	kkstdcatcher "github.com/yetiz-org/goth-kkstdcatcher"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

type SetupStdoutCatch struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupStdoutCatch) Start() {
	if !conf.Config().Logger.StdioCapture {
		return
	}

	kkstdcatcher.DefaultInstance().StdoutWriteFunc = func(s string) {
		kklogger.InfoJ("daemons:SetupStdoutCatch.Start#stdout!capture", s)
	}

	kkstdcatcher.DefaultInstance().StderrWriteFunc = func(s string) {
		kklogger.ErrorJ("daemons:SetupStdoutCatch.Start#stderr!capture", s)
	}

	kkstdcatcher.Start()
}
