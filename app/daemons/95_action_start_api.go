package daemons

import (
	"fmt"
	"net"
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	handler "github.com/yetiz-org/goth-scaffold/app/handlers"
)

type ActionStartAPI struct {
	kkdaemon.DefaultDaemon
}

func (d *ActionStartAPI) Start() {
	kklogger.InfoJ("daemons:ActionStartAPI.Start#service!start", fmt.Sprintf("Starting AppService on port %d", conf.Config().App.Port))
	handler.AppService.Start(&net.TCPAddr{Port: conf.Config().App.Port})
}

func (d *ActionStartAPI) Stop(sig os.Signal) {
	kklogger.InfoJ("daemons:ActionStartAPI.Stop#service!stop", fmt.Sprintf("Signal %s(%d)", sig, sig))
	handler.AppService.Stop()
}
