package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/yetiz-org/goth-scaffold/app/build_info"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/daemons"

	kkpanic "github.com/yetiz-org/goth-panic"
)

var (
	help       bool
	configPath string
	mode       string
)

func Initialize() {
	FlagParse()
	kkpanic.PanicNonNil(daemons.LoadActiveService())
}

func FlagParse() {
	flag.StringVar(&configPath, "c", "config.yaml", "config file")
	flag.StringVar(&mode, "m", "default", "mode: default, api, worker, db_seed, db_migration")
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
	if os.Getenv("APP_MODE") == "" && mode != "" {
		os.Setenv("APP_MODE", mode)
	}
}
