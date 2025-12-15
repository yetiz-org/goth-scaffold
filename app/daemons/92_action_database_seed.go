package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	dbseed "github.com/yetiz-org/goth-scaffold/app/database/seed"
)

type ActionDatabaseSeed struct {
	kkdaemon.DefaultDaemon
}

func (d *ActionDatabaseSeed) Start() {
	if conf.Config().DataStore.DatabaseName == "" && conf.Config().DataStore.CassandraName == "" {
		kklogger.InfoJ("daemons:ActionDatabaseSeed.Start", "database name is empty, skipping seed")
		return
	}

	if err := dbseed.RunAll(); err != nil {
		kklogger.ErrorJ("daemons:ActionDatabaseSeed.Start#run!seed", err.Error())
	}
}
