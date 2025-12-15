package errors

import (
	"fmt"

	"github.com/yetiz-org/gone/ghttp/httpstatus"

	"github.com/yetiz-org/gone/erresponse"
	"github.com/yetiz-org/gone/erresponse/constant"
	"github.com/yetiz-org/goth-kkerror"
)

var SlowDownTooFast = erresponse.Collection.Register(&erresponse.DefaultErrorResponse{
	StatusCode:  httpstatus.TooManyRequests,
	Name:        constant.ErrorSlowDown,
	Description: "too fast, rate limit exceeded",
	DefaultKKError: kkerror.DefaultKKError{
		ErrorLevel:    kkerror.Normal,
		ErrorCategory: kkerror.Client,
		ErrorCode:     "429401",
	},
})

func SlowDownTooFastWithMessage(message string) erresponse.ErrorResponse {
	return &erresponse.DefaultErrorResponse{
		StatusCode:  httpstatus.TooManyRequests,
		Name:        constant.ErrorSlowDown,
		Description: "too fast, rate limit exceeded",
		DefaultKKError: kkerror.DefaultKKError{
			ErrorLevel:    kkerror.Normal,
			ErrorCategory: kkerror.Client,
			ErrorCode:     "429401",
			ErrorMessage:  message,
		},
	}
}

// SlowDownTooFastWithFormat provides backward compatibility for dynamic format strings
func SlowDownTooFastWithFormat(format string, args ...interface{}) erresponse.ErrorResponse {
	return SlowDownTooFastWithMessage(fmt.Sprintf(format, args...))
}
