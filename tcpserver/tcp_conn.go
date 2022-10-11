package tcpserver

import (
	"context"
	"net"
	"runtime"

	"github.com/josexy/gsocks5/bufferpool"
	"github.com/josexy/logx"
)

var stackTraceBufferPool = bufferpool.NewBufferPool(func() *[]byte {
	buf := make([]byte, 2048)
	return &buf
})

type TcpConn struct {
	rwc        net.Conn
	server     *TcpServer
	remoteAddr string
}

func (conn *TcpConn) close() error {
	return conn.rwc.Close()
}

func (conn *TcpConn) serve(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			buf := stackTraceBufferPool.Get()
			n := runtime.Stack(*buf, false)
			logx.Error("%s", (*buf)[:n])
			stackTraceBufferPool.Put(buf)
		}
		if conn.server.Opts.ClientClosedHandler != nil {
			conn.server.Opts.ClientClosedHandler(conn.rwc.RemoteAddr())
		}
		logx.Warn("client closed: %s", conn.remoteAddr)
		conn.close()
	}()

	conn.remoteAddr = conn.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, conn.rwc.LocalAddr())
	logx.Info("new client incoming: %s", conn.remoteAddr)
	if conn.server.Handler != nil {
		conn.server.Handler.ServeTCP(ctx, conn.rwc)
	}
}
