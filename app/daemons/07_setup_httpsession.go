package daemons

import (
	"fmt"
	"strings"

	"github.com/yetiz-org/gone/http"
	"github.com/yetiz-org/gone/http/httpsession/redis"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

var DaemonSetupHttpSession = &SetupHttpSession{}

type SetupHttpSession struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupHttpSession) Start() {
	switch strings.ToUpper(conf.Config().Http.SessionType) {
	case string(http.SessionTypeMemory):
		http.DefaultSessionType = http.SessionTypeMemory
	case string(http.SessionTypeRedis):
		http.DefaultSessionType = http.SessionTypeRedis
		redis.RedisSessionPrefix = fmt.Sprintf("%s:%s:hs", conf.Config().Http.SessionKey, conf.Config().App.Environment)
	}

	http.SessionKey = conf.Config().Http.SessionKey
	http.SessionDomain = conf.Config().Http.SessionDomain.String()
}
