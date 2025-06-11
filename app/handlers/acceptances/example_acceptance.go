package acceptances

import (
	"fmt"
	"github.com/yetiz-org/gone/channel"

	"github.com/yetiz-org/gone/http"
)

type ExampleAcceptance struct{}

func (ExampleAcceptance) Do(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) error {
	if req.RequestURI() == "/" {
		return nil
	}

	return fmt.Errorf("error")
}
