package daemons

import (
	"os"
	"runtime"

	kkdaemon "github.com/kklab-com/goth-daemon"
	kksecret "github.com/kklab-com/goth-kksecret"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

var DaemonSetupEnvironment = &SetupEnvironment{}

type SetupEnvironment struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupEnvironment) Start() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	kksecret.PATH = conf.Config().DataStore.KKSecretPath
	os.Setenv("KKAPP_ENVIRONMENT", conf.Config().App.Environment.String())

	if conf.Config().App.Environment.Upper() != "PRODUCTION" {
		os.Setenv("KKAPP_DEBUG", "true")
	}
}
