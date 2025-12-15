package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/keyspaces"
)

type SetupKeyspaces struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupKeyspaces) Start() {
	if !keyspaces.Enabled() {
		return
	}

	if keyspaces.HealthCheck() != nil {
		kklogger.ErrorJ("daemons:SetupKeyspaces.Start#health_check!keyspaces", "keyspaces session check fail")
		panic("keyspaces session check fail")
	}
}
