package helpers

import (
	"github.com/yetiz-org/gone/ghttp"
)

type CtxChannelHttpPackHelper struct {
}

func (h *CtxChannelHttpPackHelper) UnPack(obj any) *ghttp.Pack {
	if pkg, ok := obj.(*ghttp.Pack); ok {
		return pkg
	}

	return nil
}
