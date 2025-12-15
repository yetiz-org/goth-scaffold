package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
)

type ActionShutdown struct {
	kkdaemon.DefaultDaemon
}

func (d *ActionShutdown) Start() {
	if ActiveService != nil {
		ActiveService.ShutdownGracefully()
	}
}
