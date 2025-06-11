package main

import (
	kkdaemon "github.com/yetiz-org/goth-daemon"
	"github.com/yetiz-org/goth-scaffold/app"
)

func main() {
	app.Initialize()
	if kkdaemon.Start() != nil {
		kkdaemon.ShutdownGracefully()
	}

	kkdaemon.ShutdownFuture().Await()
}
