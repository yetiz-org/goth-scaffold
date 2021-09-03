package handlers

import (
	"net"
	"os"

	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-http/http"
)

var AppService = &Service{}

type Service struct {
	ch  channel.Channel
	sig chan os.Signal
}

func (k *Service) Start(localAddr net.Addr) {
	serverBootstrap := channel.NewServerBootstrap()
	serverBootstrap.ChannelType(&http.ServerChannel{})
	serverBootstrap.ChildHandler(channel.NewInitializer(new(Initializer).Init))
	k.ch = serverBootstrap.Bind(localAddr).Sync().Channel()
}

func (k *Service) Stop() {
	k.ch.Close()
}
