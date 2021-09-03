package daemons

import (
	"fmt"
	"os"

	kkdaemon "github.com/kklab-com/goth-daemon"
	kklogger "github.com/kklab-com/goth-kklogger"
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
