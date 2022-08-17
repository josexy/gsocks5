package udpserver

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrServerClosed  = errors.New("udp: Server closed")
	ServerContextKey = &contextKey{name: "udp-server"}
)

type contextKey struct {
	name string
}

func (ck *contextKey) String() string {
	return ck.name
}

type UdpHandler interface {
	ServeUDP(context.Context, *net.UDPConn)
}

type UdpHandlerFunc func(context.Context, *net.UDPConn)

func (f UdpHandlerFunc) ServeUDP(ctx context.Context, conn *net.UDPConn) {
	f(ctx, conn)
}

type UdpServer struct {
	Addr        string
	Handler     UdpHandler
	BaseContext context.Context
	Conn        *net.UDPConn
	mu          sync.Mutex
	isShutdown  int32
	doneChan    chan struct{}
}

func NewUdpServer(addr string, handler UdpHandler) (*UdpServer, error) {
	var lAddr *net.UDPAddr
	var err error
	if addr != "" {
		if lAddr, err = net.ResolveUDPAddr("udp", addr); err != nil {
			return nil, err
		}
	}
	conn, err := net.ListenUDP("udp", lAddr)

	if err != nil {
		return nil, err
	}
	return &UdpServer{
		Addr:        conn.LocalAddr().String(),
		Handler:     handler,
		BaseContext: context.Background(),
		Conn:        conn,
		doneChan:    make(chan struct{}),
	}, nil
}

func (s *UdpServer) LocalAddr() *net.UDPAddr {
	return s.Conn.LocalAddr().(*net.UDPAddr)
}

func (s *UdpServer) IsShutdown() bool {
	return atomic.LoadInt32(&s.isShutdown) != 0
}

func (s *UdpServer) Close() error {
	if s.IsShutdown() {
		return ErrServerClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	atomic.StoreInt32(&s.isShutdown, 1)
	close(s.doneChan)
	return s.Conn.Close()
}

func (s *UdpServer) Serve() error {
	defer func() {
		recover()
		s.Close()
	}()

	ctx := context.WithValue(s.BaseContext, ServerContextKey, s)
	for {
		select {
		case <-s.getDoneChan():
			// server closed
			return ErrServerClosed
		default:
			// other error
		}
		if s.Handler != nil {
			s.Handler.ServeUDP(ctx, s.Conn)
		}
	}
}

func (s *UdpServer) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.doneChan
}
