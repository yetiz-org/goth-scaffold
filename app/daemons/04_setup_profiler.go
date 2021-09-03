package daemons

import (
	"fmt"

	"cloud.google.com/go/profiler"
	kkdaemon "github.com/kklab-com/goth-daemon"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-scaffold/app/build_info"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

var DaemonSetupProfiler = &SetupProfiler{}

type SetupProfiler struct {
	kkdaemon.DefaultDaemon
}

func (d *SetupProfiler) Start() {
	if conf.Config().Profiler.Enable {
		設定檔 := profiler.Config{
			Service:              conf.Config().App.Name.Lower(),
			ServiceVersion:       fmt.Sprintf("%s(%s-%s)", conf.Config().App.Environment, build_info.BuildGitVersion[:8], build_info.BuildTimestamp),
			MutexProfiling:       conf.Config().Profiler.MutexProfiling,
			NoAllocProfiling:     conf.Config().Profiler.NoAllocProfiling,
			NoHeapProfiling:      conf.Config().Profiler.NoHeapProfiling,
			NoGoroutineProfiling: conf.Config().Profiler.NoGoroutineProfiling,
			ProjectID:            conf.Config().Profiler.ProjectID,
		}

		if 錯誤 := profiler.Start(設定檔); 錯誤 != nil {
			kklogger.ErrorJ("SetupProfiler", map[string]interface{}{"status": "fail", "error": 錯誤.Error(), "config": 設定檔})
			panic(錯誤)
		}
	}
}
