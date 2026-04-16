package daemons

import (
	"fmt"
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/worker"
	"github.com/yetiz-org/goth-scaffold/app/worker/bootstrap"
)

type SetupWorker struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupWorker) Start() {
	namespace := fmt.Sprintf("%s-%s-asynq", conf.Config().App.Environment.Lower(), conf.Config().App.Channel.Lower())
	worker.StartClient(namespace)
	worker.RegisterService(namespace, bootstrap.RegisterTasks, bootstrap.RegisterScheduledTasks)
}

func (d *SetupWorker) Stop(sig os.Signal) {
	worker.UnRegisterService()
	worker.StopClient()
}
