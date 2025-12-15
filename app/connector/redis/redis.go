package redis

import (
	"fmt"
	"strings"
	"sync"

	datastore "github.com/yetiz-org/goth-datastore"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

var once sync.Once
var redis *datastore.Redis

func _Init() {
	once.Do(func() {
		if !Enabled() {
			return
		}

		redis = datastore.NewRedis(conf.Config().DataStore.RedisName)
	})
}

func Enabled() bool {
	return conf.Config().DataStore.RedisName != ""
}

func Instance() *datastore.Redis {
	_Init()
	return redis
}

func HealthCheck() error {
	if Master().Ping().Error != nil || Slave().Ping().Error != nil {
		return fmt.Errorf("can't connect to redis")
	}

	return nil
}

func Master() datastore.RedisOperator {
	return Instance().Master()
}

func Slave() datastore.RedisOperator {
	return Instance().Slave()
}

func Key(category string, key string) string {
	return fmt.Sprintf("%s:%s:%s:%s", conf.Config().App.Name.Upper(), conf.Config().App.Environment.Upper(), strings.ToUpper(category), key)
}
