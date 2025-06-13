package acceptances

import (
	"fmt"
	"github.com/yetiz-org/gone/channel"

	"github.com/yetiz-org/gone/ghttp"
)

type ExampleAcceptance struct{}

func (ExampleAcceptance) Do(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) error {
	if req.RequestURI() == "/" {
		return nil
	}

	return fmt.Errorf("error")
}
