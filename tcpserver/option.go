package tcpserver

import (
	"net"

	"github.com/josexy/gsocks5/logx"
)

type serverOptions struct {
	Logger              logx.Logger
	AcceptErrorHandler  func(error)
	InComingHandler     func(net.Addr)
	ClientClosedHandler func(net.Addr)
}

type ServerOption interface {
	applyTo(*serverOptions)
}

type serverOptionFunc func(*serverOptions)

func (f serverOptionFunc) applyTo(opts *serverOptions) {
	f(opts)
}

func WithLogger(logger logx.Logger) ServerOption {
	return serverOptionFunc(func(so *serverOptions) {
		so.Logger = logger
		if logger == nil {
			so.Logger = logx.DiscardLogger
		}
	})
}

func WithAcceptErrorHandler(fn func(err error)) ServerOption {
	return serverOptionFunc(func(so *serverOptions) {
		so.AcceptErrorHandler = fn
	})
}

func WithInComingHandler(fn func(addr net.Addr)) ServerOption {
	return serverOptionFunc(func(so *serverOptions) {
		so.InComingHandler = fn
	})
}

func WithClientClosedHandler(fn func(addr net.Addr)) ServerOption {
	return serverOptionFunc(func(so *serverOptions) {
		so.ClientClosedHandler = fn
	})
}
