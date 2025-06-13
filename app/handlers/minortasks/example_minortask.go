package acceptances

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
)

type ExampleMinorTask struct{}

func (ExampleMinorTask) Do(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) error {
	return nil
}
