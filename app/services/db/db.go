package db

import (
	datastore "github.com/yetiz-org/goth-datastore"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"gorm.io/gorm"
)

func Reader() *gorm.DB {
	return datastore.NewDatabase(conf.Config().DataStore.DatabaseName).Reader().DB()
}

func Writer() *gorm.DB {
	return datastore.NewDatabase(conf.Config().DataStore.DatabaseName).Writer().DB()
}
