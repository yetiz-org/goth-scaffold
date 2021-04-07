package response

import (
	"github.com/kklab-com/gone-httpstatus"
	"github.com/kklab-com/goth-erresponse"
	"github.com/kklab-com/goth-kkerror"
)

var ServerError = &erresponse.DefaultErrorResponse{
	StatusCode: httpstatus.InternalServerError,
	Name:       "server_error",
	DefaultKKError: kkerror.DefaultKKError{
		ErrorLevel:    kkerror.Urgent,
		ErrorCategory: kkerror.Server,
		ErrorCode:     "500401",
	},
}
