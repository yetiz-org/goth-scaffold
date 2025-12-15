package daemons

import (
	"fmt"

	"os"
	"time"

	"github.com/yetiz-org/goth-scaffold/app/build_info"
	"github.com/yetiz-org/goth-scaffold/app/conf"

	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	kklogger_gcp_logging "github.com/yetiz-org/goth-kklogger-gcp-logging"
	kklogger_slack "github.com/yetiz-org/goth-kklogger-slack"
)

type SetupLogger struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupLogger) Start() {
	hostname, _ := os.Hostname()
	var hooks []kklogger.LoggerHook

	if !conf.IsLocal() {
		if gcpFileBody := conf.Config().Logger.GCPLogging.JSONCredentialBody; gcpFileBody != "" {
			hooks = append(hooks,
				&kklogger_gcp_logging.KKLoggerGCPLoggingHook{
					ProjectId:       conf.Config().Profiler.ProjectID,
					LogName:         conf.Config().App.Name.String(),
					Environment:     conf.Config().App.Environment.Short(),
					CodeVersion:     build_info.BuildGitVersion,
					Service:         conf.Config().App.Name.Short(),
					ServerRoot:      hostname,
					Level:           kklogger.GetLogLevel(),
					CredentialsJSON: []byte(gcpFileBody),
				})

			kklogger.InfoJ("daemons:SetupLogger.Start#GCPLogging", map[string]interface{}{"action": "add", "status": "success"})
		}

		if conf.Config().Credentials.SecretSlack.Webhook != "" {
			hooks = append(hooks,
				&kklogger_slack.KKLoggerSlackHook{
					ServiceHookUrl: conf.Config().Credentials.SecretSlack.Webhook,
					Environment:    fmt.Sprintf("%s-%s", conf.Config().App.Name.Short(), conf.Config().App.Environment.Upper()),
					CodeVersion:    build_info.BuildGitVersion,
					ServerRoot:     hostname,
					Level:          kklogger.ErrorLevel,
				})

			kklogger.Info("daemons:SetupLogger.Start#Slack", map[string]interface{}{"action": "add", "status": "success"})
		}

	}

	kklogger.SetLoggerHooks(hooks)
	kklogger.Info("Logger Initialization")
}

func (d *SetupLogger) Stop(sig os.Signal) {
	<-time.After(time.Millisecond * 100)
}
