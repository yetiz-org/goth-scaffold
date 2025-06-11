package redis

import (
	datastore "github.com/yetiz-org/goth-kkdatastore"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

func Master() *datastore.KKRedisOp {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Master()
}

func Slave() *datastore.KKRedisOp {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Slave()
}
