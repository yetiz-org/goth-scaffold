package handlers

import (
	"strings"
	"sync"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

type Initializer struct {
	channel.DefaultInitializer
}

var initializerOnce sync.Once
var logHandler *ghttp.LogHandler
var gzipHandler, dispatchHandler channel.Handler

func (i *Initializer) Init(ch channel.Channel) {
	initializerOnce.Do(func() {
		logHandler = ghttp.NewLogHandler(strings.ToUpper(conf.Config().App.Environment.String()) != "PRODUCTION")
		gzipHandler = new(ghttp.GZipHandler)
		dispatchHandler = ghttp.NewDispatchHandler(NewRoute())
		logHandler.FilterFunc = func(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) bool {
			return req.RequestURI() != "/health"
		}
	})

	ch.Pipeline().AddLast("LOG_HANDLER", logHandler)
	ch.Pipeline().AddLast("GZIP_HANDLER", gzipHandler)
	ch.Pipeline().AddLast("DISPATCHER", dispatchHandler)
}
