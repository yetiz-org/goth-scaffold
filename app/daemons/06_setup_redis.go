package daemons

import (
	"github.com/pkg/errors"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/redis"
)

type SetupRedis struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupRedis) Start() {
	if !redis.Enabled() {
		return
	}

	datastore.DefaultRedisIdleTimeout = 60000
	datastore.DefaultRedisMaxConnLifetime = 300000
	datastore.DefaultRedisMaxIdle = 10
	datastore.DefaultRedisMaxActive = 2000
	datastore.DefaultRedisWait = true

	if redis.Master() == nil {
		panic(errors.Errorf("can't connect to master"))
	}

	if redis.Slave() == nil {
		panic(errors.Errorf("can't connect to slave"))
	}

	if err := redis.HealthCheck(); err != nil {
		kklogger.ErrorJ("daemons:SetupRedis.Start#health_check!redis", err)
		panic(err)
	}
}
