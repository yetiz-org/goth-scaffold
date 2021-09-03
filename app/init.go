package app

import (
	"flag"
	"fmt"
	"os"

	kkdaemon "github.com/kklab-com/goth-daemon"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/goth-scaffold/app/build_info"
	"github.com/kklab-com/goth-scaffold/app/conf"
	"github.com/kklab-com/goth-scaffold/app/daemons"
)

var (
	help       bool
	configPath string
)

func Initialize() {
	FlagParse()
	_RegisterDaemonService()
}

func FlagParse() {
	flag.StringVar(&configPath, "c", "config.yaml", "config file")
	flag.BoolVar(&help, "h", false, "help")
	flag.Parse()

	if help {
		fmt.Println(fmt.Sprintf("BuildGitBranch: %s", build_info.BuildGitBranch))
		fmt.Println(fmt.Sprintf("BuildGitVersion: %s", build_info.BuildGitVersion))
		fmt.Println(fmt.Sprintf("BuildGoVersion: %s", build_info.BuildGoVersion))
		fmt.Println(fmt.Sprintf("BuildTimestamp: %s", build_info.BuildTimestamp))
		flag.Usage()
		os.Exit(0)
	}

	if configPath == "" {
		println("config path can't be set empty")
		os.Exit(1)
	}

	conf.ConfigPath = configPath
}

func _RegisterDaemonService() {
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(01, daemons.DaemonSetupEnvironment))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(02, daemons.DaemonSetupLogger))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(03, daemons.DaemonSetupStdoutCatch))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(04, daemons.DaemonSetupProfiler))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(05, daemons.DaemonSetupDatabase))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(06, daemons.DaemonSetupRedis))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(07, daemons.DaemonSetupHttpSession))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(80, daemons.DaemonLoopExample))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(97, daemons.DaemonSetupUpDown))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(999, daemons.DaemonSetupLaunchService))
}
