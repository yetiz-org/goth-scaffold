package daemons

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	secret "github.com/yetiz-org/goth-secret"

	"os"
	"runtime"

	kkdaemon "github.com/yetiz-org/goth-daemon"
)

type SetupEnvironment struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupEnvironment) Start() {
	hostname, _ := os.Hostname()
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)
	secret.PATH = conf.Config().DataStore.SecretPath
	kklogger.ConfigFileName = conf.ConfigPath
	conf.Config().Logger.LoggerPath = fmt.Sprintf("%s/%d_%s-%s/", conf.Config().Logger.LoggerPath, time.Now().Unix(), time.Now().Format("2006-01-02_15:04:05"), hostname)
	os.Setenv("GOTH_LOGGER_PATH", conf.Config().Logger.LoggerPath)

	if runtime.GOOS == "linux" {
		currentDir, err := os.Getwd()
		if err == nil {
			execDir := fmt.Sprintf("%s/alloc", filepath.Dir(currentDir))
			if !strings.HasPrefix(conf.Config().Logger.LoggerPath, execDir) {
				logLinkPath := fmt.Sprintf("%s/logs", currentDir)
				os.Remove(logLinkPath)
				os.Symlink(conf.Config().Logger.LoggerPath, logLinkPath)
			}
		}
	}

	_ = os.Setenv("APP_ENVIRONMENT", conf.Config().App.Environment.Upper())

	if conf.Config().App.Environment.Upper() != "PRODUCTION" {
		_ = os.Setenv("KKAPP_DEBUG", "true")
		_ = os.Setenv("APP_DEBUG", "true")
	}
}
