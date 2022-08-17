package tcpserver

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/josexy/gsocks5/logx"
)

var (
	ErrServerClosed     = errors.New("tcp: Server closed")
	ServerContextKey    = &contextKey{name: "tcp-server"}
	LocalAddrContextKey = &contextKey{name: "tcp-addr"}
)

var defaultServerOptions = serverOptions{
	Logger: logx.DiscardLogger,
}

type contextKey struct {
	name string
}

func (ck *contextKey) String() string {
	return ck.name
}

type TcpHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}

type TcpHandlerFunc func(context.Context, net.Conn)

func (f TcpHandlerFunc) ServeTCP(ctx context.Context, conn net.Conn) {
	f(ctx, conn)
}

type TcpServer struct {
	Addr        string
	Handler     TcpHandler
	BaseContext context.Context
	Opts        serverOptions

	listener   *onceCloseListener
	mu         sync.Mutex
	isClosed   int32
	doneChan   chan struct{}
	activeConn map[*TcpConn]struct{}
}

func NewTcpServer(addr string, handler TcpHandler, opt ...ServerOption) *TcpServer {
	opts := defaultServerOptions
	for _, o := range opt {
		o.applyTo(&opts)
	}
	server := &TcpServer{
		Addr:        addr,
		Handler:     handler,
		BaseContext: context.Background(),
		Opts:        opts,
		doneChan:    make(chan struct{}),
		activeConn:  make(map[*TcpConn]struct{}),
	}

	return server
}

func (srv *TcpServer) IsClosed() bool {
	return atomic.LoadInt32(&srv.isClosed) != 0
}

func (srv *TcpServer) Close() error {
	if srv.IsClosed() {
		return ErrServerClosed
	}
	srv.mu.Lock()
	defer srv.mu.Unlock()
	atomic.StoreInt32(&srv.isClosed, 1)
	close(srv.doneChan)

	err := srv.listener.Close()
	srv.closeConns()
	return err
}

func (srv *TcpServer) closeConns() {
	for c := range srv.activeConn {
		c.close()
		delete(srv.activeConn, c)
	}
}

func (srv *TcpServer) ListenAndServe() error {
	if srv.IsClosed() {
		return ErrServerClosed
	}
	if srv.Addr == "" {
		srv.Opts.Logger.Fatal("tcp server need address")
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	srv.listener = newOnceCloseListener(ln)
	return srv.serve()
}

func (srv *TcpServer) serve() error {
	defer srv.Close()
	if srv.BaseContext == nil {
		srv.BaseContext = context.Background()
	}
	ctx := context.WithValue(srv.BaseContext, ServerContextKey, srv)
	for {
		rwc, err := srv.listener.Accept()
		if err != nil {
			select {
			case <-srv.getDoneChan():
				// server closed
				return ErrServerClosed
			default:
				// other error
			}
			if srv.Opts.AcceptErrorHandler != nil {
				srv.Opts.AcceptErrorHandler(err)
			}
			continue
		}
		if srv.Opts.InComingHandler != nil {
			srv.Opts.InComingHandler(rwc.RemoteAddr())
		}
		conn := &TcpConn{
			rwc:    rwc,
			server: srv,
		}
		srv.mu.Lock()
		srv.activeConn[conn] = struct{}{}
		srv.mu.Unlock()
		go conn.serve(ctx)
	}
}

func (srv *TcpServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}
