package daemons

import (
	"fmt"
	"os"

	kkdaemon "github.com/kklab-com/goth-daemon"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

var DaemonSetupLogger = &SetupLogger{}

type SetupLogger struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupLogger) Start() {
	// create logger path
	if _, err := os.Stat(conf.Config().Logger.LoggerPath); os.IsNotExist(err) {
		if err := os.MkdirAll(conf.Config().Logger.LoggerPath, 0755); err != nil {
			fmt.Println("logger path create fail")
			panic("logger path create fail")
		}
	}

	//hostname, _ := os.Hostname()
	kklogger.Environment = conf.Config().App.Environment.Upper()
	kklogger.LoggerPath = conf.Config().Logger.LoggerPath
	kklogger.SetLogLevel(conf.Config().Logger.LogLevel)
	kklogger.SetLoggerHooks([]kklogger.LoggerHook{
		//&kkrollbar.KKLoggerRollbarHook{
		//	Token:       conf.Config().Credentials.Rollbar.Token.String(),
		//	Environment: fmt.Sprintf("%s-%s", conf.Config().App.Name.Short(), conf.Config().App.Environment.Upper()),
		//	CodeVersion: build_info.BuildGitVersion,
		//	ServerRoot:  conf.Config().App.Name.String(),
		//	Level:       kklogger.ErrorLevel,
		//},
		//&kklogger_slack.KKLoggerSlackHook{
		//	ServiceHookUrl: "https://hooks.slack.com/services/<>",
		//	Environment:    fmt.Sprintf("%s-%s", conf.Config().App.Name.Short(), conf.Config().App.Environment.Upper()),
		//	CodeVersion:    build_info.BuildGitVersion,
		//	ServerRoot:     hostname,
		//	Level:          kklogger.ErrorLevel,
		//},
	})

	kklogger.Info("Logger Initialization")
}
