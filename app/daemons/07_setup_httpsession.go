package daemons

import (
	"fmt"
	"strings"

	redis "github.com/yetiz-org/gone-httpsession-redis"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpsession/memory"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	redis2 "github.com/yetiz-org/goth-scaffold/app/connector/redis"

	kkdaemon "github.com/yetiz-org/goth-daemon"
)

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

	redis.SessionPrefix = fmt.Sprintf("%s-%s-%s",
		conf.Config().App.Environment.Lower(),
		conf.Config().App.Channel.Lower(),
		conf.Config().App.Name.Lower())
	ghttp.SessionKey = conf.Config().App.Environment.String()
	ghttp.SessionDomain = conf.Config().Http.SessionDomain.String()
	ghttp.SessionExpireTime = conf.Config().Http.SessionExpireTime
}
