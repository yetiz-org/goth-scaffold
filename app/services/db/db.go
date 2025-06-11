package db

import (
	"github.com/jinzhu/gorm"
	datastore "github.com/yetiz-org/goth-kkdatastore"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

func Reader() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Reader().DB()
}

func Writer() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Writer().DB()
}
