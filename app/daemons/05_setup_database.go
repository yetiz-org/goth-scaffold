package daemons

import (
	kkdaemon "github.com/kklab-com/goth-daemon"
	datastore "github.com/kklab-com/goth-kkdatastore"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-scaffold/app/services/db"
)

var DaemonSetupDatabase = &SetupDatabase{}

type SetupDatabase struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupDatabase) Start() {
	datastore.KKDBParamDialTimeout = "3s"
	datastore.KKDBParamReaderMaxOpenConn = 2
	datastore.KKDBParamReaderMaxIdleConn = 1
	datastore.KKDBParamWriterMaxOpenConn = 2
	datastore.KKDBParamWriterMaxIdleConn = 1
	datastore.KKDBParamConnMaxLifetime = 60000

	if err := db.Writer().Exec("select table_name from information_schema.tables limit 1").Error; err != nil {
		kklogger.ErrorJ("daemon.SetupDatabase#Writer", err)
		panic(err)
	}

	if err := db.Reader().Exec("select table_name from information_schema.tables limit 1").Error; err != nil {
		kklogger.ErrorJ("daemon.SetupDatabase#Reader", err)
		panic(err)
	}
}
