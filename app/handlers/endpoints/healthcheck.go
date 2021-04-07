package endpoints

import (
	"bytes"

	"github.com/kklab-com/gone/http"
)

type HealthCheck struct {
	KKHandlerTask
}

var HealthOK = bytes.NewBufferString("{\"status\": \"ok\"}")

func (a *HealthCheck) Get(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.JsonResponse(HealthOK)
	return nil
}
