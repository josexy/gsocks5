package socks

import (
	"github.com/josexy/gsocks5/logx"
	"github.com/josexy/gsocks5/socks/client"
	"github.com/josexy/gsocks5/socks/server"
)

func NewSocks5Client(addr string) *client.Socks5Client {
	return client.NewSocks5Client(addr)
}

func NewSocks5Server(addr string, logger logx.Logger) *server.Socks5Server {
	return server.NewSocks5Server(addr, logger)
}
