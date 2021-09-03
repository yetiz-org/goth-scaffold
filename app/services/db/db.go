package db

import (
	"github.com/jinzhu/gorm"
	datastore "github.com/kklab-com/goth-kkdatastore"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

func Reader() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Reader().DB()
}

func Writer() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Writer().DB()
}
