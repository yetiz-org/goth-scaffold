package daemons

import (
	"net"
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/handlers"
)

var DaemonSetupLaunchService = &SetupLaunchService{}

type SetupLaunchService struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupLaunchService) Start() {
	handlers.AppService.Start(&net.TCPAddr{Port: conf.Config().App.Port})
}

func (d *SetupLaunchService) Stop(sig os.Signal) {
	handlers.AppService.Stop()
}
