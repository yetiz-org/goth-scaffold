package handlers

import (
	"net"

	"github.com/kklab-com/gone-core/channel"
	"github.com/kklab-com/gone-http/http"
)

type Service struct {
}

func (k *Service) Start(port int) {
	if port == 0 {
		port = 8080
	}

	bootstrap := channel.NewServerBootstrap()
	bootstrap.ChannelType(&http.ServerChannel{})
	bootstrap.ChildHandler(channel.NewInitializer(new(Initializer).Init))
	channel := bootstrap.Bind(&net.TCPAddr{IP: nil, Port: port}).Sync().Channel()
	channel.CloseFuture().Sync()
}
