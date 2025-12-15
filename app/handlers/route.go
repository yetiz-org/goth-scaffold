package handler

import (
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
	v1 "github.com/yetiz-org/goth-scaffold/app/handlers/endpoints/v1"
	"github.com/yetiz-org/goth-scaffold/app/handlers/minortasks"
)

type Route struct {
	ghttp.SimpleRoute
}

func NewAppRoute() *Route {
	route := Route{SimpleRoute: *ghttp.NewSimpleRoute()}
	route.SetRoot(endpoints.HandlerRoot)
	static := ghttp.NewStaticFilesHandlerTask("")
	if conf.IsDebug() {
		static.DoMinify = false
		static.DoCache = false
	}

	route.SetEndpoint("/static/*", static)
	route.SetEndpoint("/favicon.ico", static)
	route.SetEndpoint("/robots.txt", static)

	// API
	route.SetGroup("/api", minortasks.TaskDecodeSiteToken)
	route.SetGroup("/api/v1")
	route.SetEndpoint("/api/v1/health", v1.HandlerHealth)
	return &route
}
