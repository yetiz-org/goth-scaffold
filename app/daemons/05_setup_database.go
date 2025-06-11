package daemons

import (
	"github.com/pkg/errors"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	datastore "github.com/yetiz-org/goth-kkdatastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/services/db"
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

	if db.Writer() == nil {
		panic(errors.Errorf("can't connect to writer"))
	}

	if db.Reader() == nil {
		panic(errors.Errorf("can't connect to reader"))
	}

	if err := db.Writer().Exec("select table_name from information_schema.tables limit 1").Error; err != nil {
		kklogger.ErrorJ("daemon.SetupDatabase#Writer", err)
		panic(err)
	}

	if err := db.Reader().Exec("select table_name from information_schema.tables limit 1").Error; err != nil {
		kklogger.ErrorJ("daemon.SetupDatabase#Reader", err)
		panic(err)
	}
}
