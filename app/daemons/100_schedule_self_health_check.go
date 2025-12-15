package daemons

import (
	"os"

	"github.com/yetiz-org/gone/erresponse"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/connector/keyspaces"
	"github.com/yetiz-org/goth-scaffold/app/connector/redis"
)

type ScheduleSelfHealthCheck struct {
	kkdaemon.DefaultSchedulerDaemon
}

func (d *ScheduleSelfHealthCheck) Start() {
}

func (d *ScheduleSelfHealthCheck) Stop(sig os.Signal) {
}

func (d *ScheduleSelfHealthCheck) When() kkdaemon.CronSyntax {
	return "* * * * *"
}

func (d *ScheduleSelfHealthCheck) Loop() error {
	if redis.Enabled() {
		if err := redis.HealthCheck(); err != nil {
			kklogger.ErrorJ("daemons:ScheduleSelfHealthCheck.Loop#HealthCheck!redis", "redis session check fail")
			d.shutdownGracefully()
			return erresponse.ServerErrorCrossServiceOperationFail
		}
	}

	if database.Enabled() {
		if err := database.HealthCheck(); err != nil {
			kklogger.ErrorJ("daemons:ScheduleSelfHealthCheck.Loop#HealthCheck!datastore", "datastore session check fail")
			d.shutdownGracefully()
			return erresponse.ServerErrorCrossServiceOperationFail
		}
	}

	if keyspaces.Enabled() {
		if err := keyspaces.HealthCheck(); err != nil {
			kklogger.ErrorJ("daemons:ScheduleSelfHealthCheck.Loop#HealthCheck!keyspaces", "keyspaces session check fail")
			d.shutdownGracefully()
			return erresponse.ServerErrorCrossServiceOperationFail
		}
	}

	return nil
}

func (d *ScheduleSelfHealthCheck) shutdownGracefully() {
	go func() {
		if ActiveService != nil {
			ActiveService.ShutdownGracefully()
		}
	}()
}
