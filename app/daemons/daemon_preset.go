package daemons

import (
	"os"

	kkdaemon "github.com/yetiz-org/goth-daemon"
)

var ActiveService *kkdaemon.DaemonService

type daemonEntry struct {
	new   func() kkdaemon.Daemon
	order int
}

var presetMaps = map[string][]daemonEntry{}

func init() {
	presetMaps["default"] = []daemonEntry{
		{new: func() kkdaemon.Daemon { return &SetupEnvironment{} }, order: 1},
		{new: func() kkdaemon.Daemon { return &SetupLogger{} }, order: 2},
		{new: func() kkdaemon.Daemon { return &SetupStdoutCatch{} }, order: 3},
		{new: func() kkdaemon.Daemon { return &SetupProfiler{} }, order: 4},
		{new: func() kkdaemon.Daemon { return &SetupDatabase{} }, order: 5},
		{new: func() kkdaemon.Daemon { return &SetupRedis{} }, order: 6},
		{new: func() kkdaemon.Daemon { return &SetupHttpSession{} }, order: 7},
		{new: func() kkdaemon.Daemon { return &SetupKeyspaces{} }, order: 10},
		{new: func() kkdaemon.Daemon { return &SchedulePerformanceMeasure{} }, order: 22},
		{new: func() kkdaemon.Daemon { return &ActionDatabaseMigration{} }, order: 91},
		{new: func() kkdaemon.Daemon { return &ActionStartAPI{} }, order: 95},
		{new: func() kkdaemon.Daemon { return &ScheduleSelfHealthCheck{} }, order: 100},
	}
}

func NewDaemonServiceForMode(mode string) (*kkdaemon.DaemonService, error) {
	entries, found := presetMaps[mode]
	if !found {
		entries = presetMaps["default"]
	}

	s := kkdaemon.NewDaemonService()
	for _, entry := range entries {
		if err := s.RegisterDaemonWithOrder(entry.new(), entry.order); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func LoadActiveService() error {
	svc, err := NewDaemonServiceForMode(os.Getenv("APP_MODE"))
	if err != nil {
		return err
	}

	ActiveService = svc
	return nil
}
