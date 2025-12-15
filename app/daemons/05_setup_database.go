package daemons

import (
	"github.com/pkg/errors"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
)

type SetupDatabase struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupDatabase) Start() {
	if !database.Enabled() {
		return
	}

	datastore.DefaultDatabaseDialTimeout = "3s"
	datastore.DefaultDatabaseMaxOpenConn = 10
	datastore.DefaultDatabaseMaxIdleConn = 2
	datastore.DefaultDatabaseMaxOpenConn = 10
	datastore.DefaultDatabaseMaxIdleConn = 2
	datastore.DefaultDatabaseConnMaxLifetime = 60000

	if database.Writer() == nil {
		panic(errors.Errorf("can't connect to writer db"))
	}

	if database.Reader() == nil {
		panic(errors.Errorf("can't connect to reader db"))
	}

	if err := database.HealthCheck(); err != nil {
		kklogger.ErrorJ("daemons:SetupDatabase.Start#health_check!database", err)
		panic(err)
	}
}
