package daemons

import (
	"fmt"

	"cloud.google.com/go/profiler"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/build_info"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"google.golang.org/api/option"
)

type SetupProfiler struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupProfiler) Start() {
	if conf.Config().Profiler.Enable && conf.Config().Profiler.JSONCredentialBody != "" {
		config := profiler.Config{
			Service:              conf.Config().App.Name.Lower(),
			ServiceVersion:       fmt.Sprintf("%s(%s-%s)", conf.Config().App.Environment, build_info.BuildGitVersion[:8], build_info.BuildTimestamp),
			MutexProfiling:       conf.Config().Profiler.MutexProfiling,
			NoAllocProfiling:     conf.Config().Profiler.NoAllocProfiling,
			NoHeapProfiling:      conf.Config().Profiler.NoHeapProfiling,
			NoGoroutineProfiling: conf.Config().Profiler.NoGoroutineProfiling,
			ProjectID:            conf.Config().Profiler.ProjectID,
		}

		if err := profiler.Start(config, option.WithCredentialsJSON([]byte(conf.Config().Profiler.JSONCredentialBody))); err != nil {
			kklogger.ErrorJ("daemons:SetupProfiler.Start#profiler", map[string]interface{}{"status": "fail", "error": err.Error(), "config": config})
			panic(err)
		}
	}
}
