package handlers

import (
	"github.com/yetiz-org/gone/http"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
)

type Route struct {
	http.DefaultRoute
}

func NewRoute() *Route {
	route := Route{DefaultRoute: *http.NewRoute()}
	route.
		SetRoot(http.NewEndPoint("", endpoints.HandlerRoot, nil)).
		AddRecursivePoint(http.NewEndPoint("r", endpoints.HandlerRoot, nil)).
		AddRecursivePoint(http.NewEndPoint("static", http.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(http.NewEndPoint("favicon.ico", http.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(http.NewEndPoint("robots.txt", http.NewStaticFilesHandlerTask(""), nil)).
		AddEndPoint(http.NewEndPoint("health", new(endpoints.HealthCheck), nil))
	return &route
}
