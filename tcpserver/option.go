package tcpserver

import (
	"net"
)

type serverOptions struct {
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
