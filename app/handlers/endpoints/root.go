package endpoints

import (
	"github.com/kklab-com/gone/http"
)

type Root struct {
	KKHandlerTask
}

var HandlerRoot = &Root{}

func (l *Root) Get(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	l.RenderHtml("home", nil, resp)
	return nil
}
