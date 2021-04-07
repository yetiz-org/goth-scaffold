package main

import (
	"github.com/kklab-com/goth-scaffold/app"
	"github.com/kklab-com/goth-scaffold/app/conf"
	"github.com/kklab-com/goth-scaffold/app/handlers"
)

func main() {
	startService()
}

func startService() {
	app.Init()
	service := handlers.Service{}
	service.Start(conf.Config().App.Port)
}
