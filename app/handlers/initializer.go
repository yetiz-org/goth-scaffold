package handler

import (
	"mime"
	"sync"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpheadername"
	"github.com/yetiz-org/gone/gws"
	buf "github.com/yetiz-org/goth-bytebuf"
	kklogger "github.com/yetiz-org/goth-kklogger"
	kktemplate "github.com/yetiz-org/goth-kktemplate"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
)

type Initializer struct {
	channel.DefaultInitializer
}

var initializerOnce sync.Once
var trackHandler = &TrackHandler{}
var logHandler *ghttp.LogHandler
var gzipHandler channel.Handler
var dispatcher *ghttp.DispatchHandler

func (i *Initializer) Init(ch channel.Channel) {
	initializerOnce.Do(func() {
		logHandler = ghttp.NewLogHandler(conf.Config().App.Environment.Upper() != "PRODUCTION")
		gzipHandler = &ghttp.GZipHandler{
			CompressThreshold: 1024,
		}

		dispatcher = ghttp.NewDispatchHandler(NewAppRoute())

		// Set custom 404 handler
		dispatcher.DefaultStatusResponse[404] = func(req *ghttp.Request, resp *ghttp.Response, params map[string]any) {
			render404Page(req, resp)
		}

		logHandler.FilterFunc = func(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) bool {
			return req.RequestURI() != "/api/v1/health"
		}
	})

	ch.Pipeline().AddLast("LOG_HANDLER", logHandler)
	ch.Pipeline().AddLast("GZIP_HANDLER", gzipHandler)
	ch.Pipeline().AddLast("TRACK_HANDLER", trackHandler)
	ch.Pipeline().AddLast("DISPATCHER", dispatcher)
	ch.Pipeline().AddLast("WS_UPGRADER", &gws.UpgradeProcessor{})
}

// render404Page renders the custom 404 error page
func render404Page(req *ghttp.Request, resp *ghttp.Response) {
	// Get language preference
	handlerTask := &endpoints.HandlerTask{}
	lang := handlerTask.Lang(req)

	// Load and render 404 template
	if tmpl, err := kktemplate.LoadFrameHtml("_404", lang); err == nil {
		buffer := buf.EmptyByteBuf()
		renderVars := map[string]interface{}{
			"PageID": "404",
		}

		if e := tmpl.ExecuteTemplate(buffer, "main", renderVars); e != nil {
			kklogger.ErrorJ("handler:render404Page#template!execute_fail", e.Error())
			// Fallback to simple HTML if template execution fails
			resp.SetBody(buf.NewByteBufString("<html><body><h1>404 - Page Not Found</h1></body></html>"))
		} else {
			resp.SetHeader(httpheadername.ContentType, mime.TypeByExtension(".html"))
			resp.SetBody(buffer)
		}
	} else {
		kklogger.ErrorJ("handler:render404Page#template!load_fail", err.Error())
		// Fallback to simple HTML if template loading fails
		resp.SetBody(buf.NewByteBufString("<html><body><h1>404 - Page Not Found</h1></body></html>"))
	}
}
