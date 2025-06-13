package handlers

import (
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
)

type Route struct {
	ghttp.DefaultRoute
}

func NewRoute() *Route {
	route := Route{DefaultRoute: *ghttp.NewRoute()}
	route.
		SetRoot(ghttp.NewEndPoint("", endpoints.HandlerRoot, nil)).
		AddRecursivePoint(ghttp.NewEndPoint("r", endpoints.HandlerRoot, nil)).
		AddRecursivePoint(ghttp.NewEndPoint("static", ghttp.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(ghttp.NewEndPoint("favicon.ico", ghttp.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(ghttp.NewEndPoint("robots.txt", ghttp.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(ghttp.NewEndPoint("health", new(endpoints.HealthCheck), nil))
	return &route
}
