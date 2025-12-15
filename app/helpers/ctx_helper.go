package helpers

import (
	"context"

	"github.com/yetiz-org/gone/channel"
)

type CtxHelper struct {
}

func (h *CtxHelper) GetCtxValue(ctx context.Context, key channel.ParamKey) any {
	return ctx.Value(key)
}

func (h *CtxHelper) AttachCtxValue(ctx context.Context, key channel.ParamKey, value any) context.Context {
	return context.WithValue(ctx, key, value)
}
