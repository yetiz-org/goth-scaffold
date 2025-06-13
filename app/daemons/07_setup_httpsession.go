package daemons

import (
	"fmt"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpsession/redis"
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
	case string(ghttp.SessionTypeMemory):
		ghttp.DefaultSessionType = ghttp.SessionTypeMemory
	case string(ghttp.SessionTypeRedis):
		ghttp.DefaultSessionType = ghttp.SessionTypeRedis
		redis.RedisSessionPrefix = fmt.Sprintf("%s:%s:hs", conf.Config().Http.SessionKey, conf.Config().App.Environment)
	}

	ghttp.SessionKey = conf.Config().Http.SessionKey
	ghttp.SessionDomain = conf.Config().Http.SessionDomain.String()
}
