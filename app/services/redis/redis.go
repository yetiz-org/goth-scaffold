package redis

import (
	datastore "github.com/kklab-com/goth-kkdatastore"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

func Master() *datastore.KKRedisOp {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Master()
}

func Slave() *datastore.KKRedisOp {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Slave()
}
