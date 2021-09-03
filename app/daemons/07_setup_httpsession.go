package daemons

import (
	"fmt"
	"strings"

	"github.com/kklab-com/gone-http/http"
	"github.com/kklab-com/gone-http/http/httpsession/redis"
	kkdaemon "github.com/kklab-com/goth-daemon"
	"github.com/kklab-com/goth-scaffold/app/conf"
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
