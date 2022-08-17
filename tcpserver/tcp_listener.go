package tcpserver

import (
	"net"
	"sync"
)

type onceCloseListener struct {
	net.Listener
	once     sync.Once
	closeErr error
}

func newOnceCloseListener(ln net.Listener) *onceCloseListener {
	return &onceCloseListener{
		Listener: ln,
	}
}

func (ln *onceCloseListener) Close() error {
	ln.once.Do(func() {
		ln.closeErr = ln.Listener.Close()
	})
	return ln.closeErr
}

func (ln *onceCloseListener) Accept() (net.Conn, error) {
	return ln.Listener.Accept()
}
