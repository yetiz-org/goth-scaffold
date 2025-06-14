package redis

import (
	datastore "github.com/yetiz-org/goth-datastore"
	"sync"
)

var once sync.Once
var redis *datastore.Redis

func Init(profileName string) {
	once.Do(func() {
		redis = datastore.NewRedis(profileName)
	})
}

func Instance() *datastore.Redis {
	return redis
}

func Master() *datastore.RedisOp {
	return redis.Master()
}

func Slave() *datastore.RedisOp {
	return redis.Slave()
}
