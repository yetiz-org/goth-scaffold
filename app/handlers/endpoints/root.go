package endpoints

import (
	"time"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
)

type Root struct {
	HandlerTask
}

var HandlerRoot = &Root{}

func (h *Root) Index(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	req.Session().PutInt64("COUNT", req.Session().GetInt64("COUNT")+1)

	h.RenderHtml("home", &RenderConfig{
		PageTitle: "the root title",
		PageRenderData: map[string]interface{}{
			"SHOW":      "SHOW~~",
			"CSRFToken": h.GenerateCSRFToken([]byte(req.Session().Id()), time.Hour),
		},
	}, resp)
	return nil
}
