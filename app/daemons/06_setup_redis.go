package daemons

import (
	"fmt"
	"github.com/pkg/errors"
	redis "github.com/yetiz-org/gone-httpsession-redis"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	redis2 "github.com/yetiz-org/goth-scaffold/app/services/redis"
)

var DaemonSetupRedis = &SetupRedis{}

type SetupRedis struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupRedis) Start() {
	redis.SessionPrefix = fmt.Sprintf("%s:%s", conf.Config().Http.SessionKey, conf.Config().App.Environment)
	datastore.DefaultRedisIdleTimeout = 60000
	datastore.DefaultRedisMaxConnLifetime = 300000
	datastore.DefaultRedisMaxIdle = 10
	datastore.DefaultRedisMaxActive = 2000
	datastore.DefaultRedisWait = true

	redis2.Init(conf.Config().DataStore.RedisName)
	if redis2.Master() == nil {
		panic(errors.Errorf("can't connect to master"))
	}

	if redis2.Slave() == nil {
		panic(errors.Errorf("can't connect to slave"))
	}

	if _, err := redis2.Master().Conn().Do("PING"); err != nil {
		kklogger.ErrorJ("daemon.SetupRedis#Master", err)
		panic(err)
	}

	if _, err := redis2.Slave().Conn().Do("PING"); err != nil {
		kklogger.ErrorJ("daemon.SetupRedis#Slave", err)
		panic(err)
	}
}
