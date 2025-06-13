package endpoints

import (
	"bytes"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
)

type HealthCheck struct {
	HandlerTask
}

var HealthOK = bytes.NewBufferString("{\"status\": \"ok\"}")

func (a *HealthCheck) Get(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.JsonResponse(HealthOK)
	return nil
}
