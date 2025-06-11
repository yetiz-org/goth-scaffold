package acceptances

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/http"
)

type ExampleMinorTask struct{}

func (ExampleMinorTask) Do(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) error {
	return nil
}
