// Defines the public API health-check handler and its documented response shape.

package v1

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	buf "github.com/yetiz-org/goth-bytebuf"
	"github.com/yetiz-org/goth-scaffold/app/handlers/endpoints"
)

// Health serves the public liveness endpoint.
// @goai.tag Public
// @goai.security none
type Health struct {
	endpoints.HandlerTask
}

// HandlerHealth is the route-owned health handler instance.
var HandlerHealth = &Health{}

// HealthOK is the shared serialized healthy response body.
var HealthOK = buf.NewByteBufString("{\"status\": \"ok\"}")

// HealthResponse documents the JSON object returned by the health endpoint.
// @goai.schemaName HealthResponse
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// Get GET /api/v1/health
// Returns a JSON object whose status is "ok" when the service is healthy.
// @goai.endpoint GET /api/v1/health
// @goai.summary Check service health
// @goai.description Returns the current liveness status without authentication.
// @goai.operationId health.get
// @goai.response 200 "Service is healthy." mediaType=application/json schemaType=HealthResponse
// @goai.example response 200 application/json healthy "Healthy response" {"status":"ok"}
func (a *Health) Get(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) (errorResponse ghttp.ErrorResponse) {
	resp.JsonResponse(HealthOK)
	return nil
}
