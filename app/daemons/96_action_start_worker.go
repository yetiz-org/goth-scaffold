package daemons

import (
	"fmt"
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/worker"
)

type ActionStartWorker struct {
	kkdaemon.DefaultDaemon
}

func (d *ActionStartWorker) Start() {
	qname := fmt.Sprintf("%s-%s-asynq", conf.Config().App.Environment.Lower(), conf.Config().App.Channel.Lower())
	kklogger.InfoJ("daemons:ActionStartWorker.Start#service!start", fmt.Sprintf("Starting Worker %s", qname))
	worker.StartService(qname)
}

func (d *ActionStartWorker) Stop(sig os.Signal) {
	kklogger.InfoJ("daemons:ActionStartWorker.Stop#service!stop", fmt.Sprintf("Signal %s(%d)", sig, sig))
	worker.StopService()
}
