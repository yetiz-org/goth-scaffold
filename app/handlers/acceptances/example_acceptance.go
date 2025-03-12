package acceptances

import (
	"fmt"
	"github.com/kklab-com/gone-core/channel"

	"github.com/kklab-com/gone-http/http"
)

type ExampleAcceptance struct{}

func (ExampleAcceptance) Do(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) error {
	if req.RequestURI() == "/" {
		return nil
	}

	return fmt.Errorf("error")
}
