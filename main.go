package main

import (
	kkdaemon "github.com/kklab-com/goth-daemon"
	"github.com/kklab-com/goth-scaffold/app"
)

func main() {
	app.Initialize()
	kkdaemon.Start()
	kkdaemon.WaitShutdown()
}
