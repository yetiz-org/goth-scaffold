package endpoints

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
)

type Root struct {
	HandlerTask
}

var HandlerRoot = &Root{}

func (l *Root) Get(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	l.SessionRenderData(resp, "SSS", "S1")
	l.RenderHtml("home", &RenderConfig{
		PageTitle:      "the root title",
		PageRenderData: map[string]interface{}{"SHOW": "SHOW~~"},
	}, resp)
	return nil
}
