package handler

import (
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/goth-scaffold/app/helpers"
)

type TrackHandler struct {
	channel.DefaultHandler
	helpers.CtxChannelHttpPackHelper
}

func (h *TrackHandler) Write(ctx channel.HandlerContext, obj any, future channel.Future) {
	if pack := h.UnPack(obj); pack != nil {
		pack.Response.Header().Set("x-gone-channel-id", pack.Request.Channel().ID())
		pack.Response.Header().Set("x-gone-track-id", pack.Request.TrackID())
	}

	ctx.Write(obj, future)
}
