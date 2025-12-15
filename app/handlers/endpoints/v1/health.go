package v1

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	buf "github.com/yetiz-org/goth-bytebuf"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
)

type Health struct {
	endpoints.HandlerTask
}

var HandlerHealth = &Health{}

var HealthOK = buf.NewByteBufString("{\"status\": \"ok\"}")

// Get GET /api/v1/health
// Health check endpoint. Returns status "ok" when service is healthy.
func (a *Health) Get(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.JsonResponse(HealthOK)
	return nil
}
