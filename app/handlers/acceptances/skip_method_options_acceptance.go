package acceptances

import "github.com/yetiz-org/gone/ghttp"

type SkipMethodOptionsAcceptance struct {
	*ghttp.DispatchAcceptance
}

func (a *SkipMethodOptionsAcceptance) SkipMethodOptions() bool {
	return true
}
