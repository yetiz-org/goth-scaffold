package handler

import (
	"net"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/gws"
)

var AppService = &Service{}

type Service struct {
	ch channel.Channel
}

func (k *Service) Start(localAddr net.Addr) {
	initializer := &Initializer{}
	serverBootstrap := channel.NewServerBootstrap()
	serverBootstrap.ChannelType(&ghttp.ServerChannel{})
	serverBootstrap.SetParams(ghttp.ParamIdleTimeout, 600)
	serverBootstrap.SetParams(ghttp.ParamReadTimeout, 600)
	serverBootstrap.SetParams(ghttp.ParamWriteTimeout, 600)
	serverBootstrap.SetParams(ghttp.ParamMaxBodyBytes, 30<<20)
	serverBootstrap.ChildHandler(channel.NewInitializer(initializer.Init))
	serverBootstrap.SetChildParams(gws.ParamCheckOrigin, false)
	k.ch = serverBootstrap.Bind(localAddr).Sync().Channel()
}

func (k *Service) Stop() {
	k.ch.Close()
}
