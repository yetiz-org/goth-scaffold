package daemons

import (
	"fmt"
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

var DaemonSetupUpDown = &SetupUpDown{}

type SetupUpDown struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupUpDown) Start() {
	hostname, _ := os.Hostname()
	kklogger.InfoJ("SetupUpDown", fmt.Sprintf("%s up", hostname))
}

func (d *SetupUpDown) Stop(sig os.Signal) {
	hostname, _ := os.Hostname()
	kklogger.InfoJ("SetupUpDown", fmt.Sprintf("%s down", hostname))
}
