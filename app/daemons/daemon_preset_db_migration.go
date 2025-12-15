package daemons

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
)

func init() {
	presetMaps["db_migration"] = []daemonEntry{
		{new: func() kkdaemon.Daemon { return &SetupEnvironment{} }, order: 1},
		{new: func() kkdaemon.Daemon { return &SetupLogger{} }, order: 2},
		{new: func() kkdaemon.Daemon { return &SetupStdoutCatch{} }, order: 3},
		{new: func() kkdaemon.Daemon { return &SetupProfiler{} }, order: 4},
		{new: func() kkdaemon.Daemon { return &SetupDatabase{} }, order: 5},
		{new: func() kkdaemon.Daemon { return &SetupRedis{} }, order: 6},
		{new: func() kkdaemon.Daemon { return &SetupHttpSession{} }, order: 7},
		{new: func() kkdaemon.Daemon { return &SetupKeyspaces{} }, order: 10},
		{new: func() kkdaemon.Daemon { return &ActionDatabaseMigration{} }, order: 91},
		{new: func() kkdaemon.Daemon { return &ActionShutdown{} }, order: 9999},
	}
}
