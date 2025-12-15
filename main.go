package main

import (
	"github.com/yetiz-org/goth-scaffold/app"
	"github.com/yetiz-org/goth-scaffold/app/daemons"
)

func main() {
	app.Initialize()
	if daemons.ActiveService.Start() != nil {
		daemons.ActiveService.ShutdownGracefully()
	}

	daemons.ActiveService.ShutdownFuture().Await()
}
