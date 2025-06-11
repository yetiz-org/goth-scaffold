package endpoints

import (
	"bytes"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/http"
)

type HealthCheck struct {
	HandlerTask
}

var HealthOK = bytes.NewBufferString("{\"status\": \"ok\"}")

func (a *HealthCheck) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.JsonResponse(HealthOK)
	return nil
}
