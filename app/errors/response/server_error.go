package response

import (
	"github.com/yetiz-org/gone/erresponse"
	"github.com/yetiz-org/gone/ghttp/httpstatus"
	kkerror "github.com/yetiz-org/goth-kkerror"
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
