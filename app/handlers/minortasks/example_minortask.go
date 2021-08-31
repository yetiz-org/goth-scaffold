package acceptances

import (
	"github.com/kklab-com/gone-http/http"
)

type ExampleMinorTask struct{}

func (ExampleMinorTask) Do(req *http.Request, resp *http.Response, params map[string]interface{}) error {
	return nil
}
