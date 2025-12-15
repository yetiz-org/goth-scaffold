package database

import (
	"sync"

	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"gorm.io/gorm"
)

var once sync.Once
var database *datastore.Database

func _Init() {
	once.Do(func() {
		if !Enabled() {
			return
		}

		database = datastore.NewDatabase(conf.Config().DataStore.DatabaseName)
	})
}

func Enabled() bool {
	return conf.Config().DataStore.DatabaseName != ""
}

func Instance() *datastore.Database {
	_Init()
	return database
}

func HealthCheck() error {
	db, err := Writer().DB()
	if err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Writer!health_check", err)
		return err
	}

	if err := db.Ping(); err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Writer!health_check", err)
		return err
	}

	db, err = Reader().DB()
	if err != nil {
		kklogger.ErrorJ("database:Database.HealthCheck#Reader!health_check", err)
		return err
	}

	if err := db.Ping(); err != nil {
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
