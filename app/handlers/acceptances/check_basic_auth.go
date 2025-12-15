package acceptances

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/erresponse"
	"github.com/yetiz-org/gone/ghttp"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

type CheckBasicAuth struct {
	SkipMethodOptionsAcceptance
}

var HCheckBasicAuth = &CheckBasicAuth{}

func (a *CheckBasicAuth) Do(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) (er error) {
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		kklogger.WarnJ("acceptances:CheckBasicAuth.Do#auth!missing_header", map[string]any{"requestURI": req.RequestURI()})
		resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.Config().App.Name.String()))
		resp.ResponseError(erresponse.InvalidToken)
		return erresponse.InvalidToken
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		kklogger.WarnJ("acceptances:CheckBasicAuth.Do#auth!invalid_scheme", map[string]any{"authHeader": authHeader})
		resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.Config().App.Name.String()))
		resp.ResponseError(erresponse.InvalidToken)
		return erresponse.InvalidToken
	}

	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		kklogger.WarnJ("acceptances:CheckBasicAuth.Do#auth!decode_error", err.Error())
		resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.Config().App.Name.String()))
		resp.ResponseError(erresponse.InvalidToken)
		return erresponse.InvalidToken
	}

	credentials := string(decodedBytes)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		kklogger.WarnJ("acceptances:CheckBasicAuth.Do#auth!invalid_format", map[string]any{"credentials": credentials})
		resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.Config().App.Name.String()))
		resp.ResponseError(erresponse.InvalidToken)
		return erresponse.InvalidToken
	}

	username := parts[0]
	password := parts[1]

	if username != "goth" || password != "scaffold" {
		kklogger.WarnJ("acceptances:CheckBasicAuth.Do#auth!invalid_credentials", map[string]any{"username": username})
		resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, conf.Config().App.Name.String()))
		resp.ResponseError(erresponse.InvalidToken)
		return erresponse.InvalidToken
	}

	return
}
