package endpoints

import (
	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/http"
)

type Root struct {
	HandlerTask
}

var HandlerRoot = &Root{}

func (l *Root) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	l.SessionRenderData(resp, "SSS", "S1")
	l.RenderHtml("home", &RenderConfig{
		PageTitle:      "the root title",
		PageRenderData: map[string]interface{}{"SHOW": "SHOW~~"},
	}, resp)
	return nil
}
