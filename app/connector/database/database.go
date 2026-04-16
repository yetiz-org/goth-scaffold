package database

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type _IgnoreRecordNotFoundLogger struct {
	gormlogger.Interface
}

func (l _IgnoreRecordNotFoundLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	l.Interface.Trace(ctx, begin, fc, err)
}

func _ApplyReadCommittedIsolation(op datastore.DatabaseOperator) {
	if op == nil {
		return
	}

	params := op.GetConnParams()
	params.TransactionIsolation = datastore.DatabaseIsolationLevelReadCommitted
	op.SetConnParams(params)
}

func _ApplyIgnoreRecordNotFoundLogger(op datastore.DatabaseOperator) {
	if op == nil {
		return
	}

	if baseLogger := op.GetLogger(); baseLogger != nil {
		op.SetLogger(_IgnoreRecordNotFoundLogger{Interface: baseLogger})
		return
	}

	cfg := op.GetGORMParams()
	if cfg.Logger != nil {
		cfg.Logger = _IgnoreRecordNotFoundLogger{Interface: cfg.Logger}
		op.SetGORMParams(cfg)
		return
	}

	op.SetLogger(_IgnoreRecordNotFoundLogger{Interface: gormlogger.Default})
}

var once sync.Once
var db *datastore.Database

func _Init() {
	once.Do(func() {
		if !Enabled() {
			return
		}

		db = datastore.NewDatabase(conf.Config().DataStore.DatabaseName)
		if db == nil {
			return
		}

		_ApplyIgnoreRecordNotFoundLogger(db.Reader())
		_ApplyIgnoreRecordNotFoundLogger(db.Writer())
		_ApplyReadCommittedIsolation(db.Reader())
	})
}

func Enabled() bool {
	return conf.Config().DataStore.DatabaseName != ""
}

func Instance() *datastore.Database {
	_Init()
	return db
}

func HealthCheck() error {
	if Instance() == nil {
		return fmt.Errorf("database not initialized: check secret_path and database_name configuration")
	}

	sqlDB, err := Writer().DB()
	if err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Writer!health_check", err)
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Writer!health_check", err)
		return err
	}

	sqlDB, err = Reader().DB()
	if err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Reader!health_check", err)
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Reader!health_check", err)
		return err
	}

	return nil
}

func Reader() *gorm.DB {
	return Instance().Reader().DB()
}

func Writer() *gorm.DB {
	return Instance().Writer().DB()
}
