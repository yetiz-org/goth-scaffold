package app

import (
	"flag"
	"fmt"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	"os"

	kkpanic "github.com/yetiz-org/goth-panic"
	"github.com/yetiz-org/goth-scaffold/app/build_info"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/daemons"
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
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupEnvironment))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupLogger))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupStdoutCatch))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupProfiler))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupDatabase))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupRedis))
	//kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupHttpSession))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonTimerLoopExample))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSchedulerLoopExample))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupUpDown))
	kkpanic.PanicNonNil(kkdaemon.RegisterDaemon(daemons.DaemonSetupLaunchService))
}
