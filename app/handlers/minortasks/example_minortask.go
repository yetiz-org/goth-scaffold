package acceptances

import (
	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-http/http"
)

type ExampleMinorTask struct{}

func (ExampleMinorTask) Do(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) error {
	return nil
}
