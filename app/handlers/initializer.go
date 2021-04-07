package handlers

import (
	"strings"
	"sync"

	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/http"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

type Initializer struct {
	channel.DefaultInitializer
}

var initializerOnce sync.Once
var logHandler *http.LogHandler
var gzipHandler, dispatchHandler channel.Handler

func (i *Initializer) Init(ch channel.Channel) {
	initializerOnce.Do(func() {
		logHandler = http.NewLogHandler(strings.ToUpper(conf.Config().App.Environment.String()) != "PRODUCTION")
		gzipHandler = new(http.GZipHandler)
		dispatchHandler = http.NewDispatchHandler(NewRoute())
		logHandler.FilterFunc = func(req *http.Request, resp *http.Response, params map[string]interface{}) bool {
			return req.RequestURI != "/health"
		}
	})

	ch.Pipeline().AddLast("LOG_HANDLER", logHandler)
	ch.Pipeline().AddLast("GZIP_HANDLER", gzipHandler)
	ch.Pipeline().AddLast("DISPATCHER", dispatchHandler)
}
