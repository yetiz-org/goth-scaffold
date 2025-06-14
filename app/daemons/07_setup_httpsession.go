package daemons

import (
	redis "github.com/yetiz-org/gone-httpsession-redis"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpsession/memory"
	redis2 "github.com/yetiz-org/goth-scaffold/app/services/redis"
	"strings"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

var DaemonSetupHttpSession = &SetupHttpSession{}

type SetupHttpSession struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupHttpSession) Start() {
	switch strings.ToUpper(conf.Config().Http.SessionType) {
	case string(memory.SessionTypeMemory):
		ghttp.DefaultSessionType = memory.SessionTypeMemory
	case string(redis.SessionTypeRedis):
		ghttp.DefaultSessionType = redis.SessionTypeRedis
		ghttp.RegisterSessionProvider(redis.NewSessionProviderWithRedis(redis2.Instance()))
	}

	ghttp.SessionKey = conf.Config().Http.SessionKey
	ghttp.SessionDomain = conf.Config().Http.SessionDomain.String()
}
