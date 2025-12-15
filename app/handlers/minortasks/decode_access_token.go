package minortasks

import (
	"strings"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpheadername"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/constant/param"
	"github.com/yetiz-org/goth-scaffold/app/helpers"
)

type DecodeSiteToken struct {
	*ghttp.DispatchAcceptance
	helpers.ParamsHelper
}

var TaskDecodeSiteToken = &DecodeSiteToken{}

func (h *DecodeSiteToken) Do(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) (er error) {
	var tokenString string
	// get token from header
	if authSp := strings.Split(req.Header().Get(httpheadername.Authorization), " "); len(authSp) == 2 && strings.ToUpper(authSp[0]) == "BEARER" {
		tokenString = authSp[1]
	}

	// get token from websocket protocol
	if v := req.Header().Get("Sec-WebSocket-Protocol"); v != "" {
		if protocols := strings.Split(v, ", "); len(protocols) > 1 {
			if strings.ToUpper(protocols[0]) == conf.Config().App.Name.Upper() {
				resp.AddHeader("Sec-WebSocket-Protocol", protocols[0])
				tokenString = protocols[1]
			}
		}
	}

	if tokenString != "" {
		h.SetParam(params, param.AccessTokenString, tokenString)
		return nil
	}

	return nil
}
