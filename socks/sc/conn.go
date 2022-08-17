package sc

import "net"

type Conn interface {
	net.Conn
	TCP() *net.TCPConn
	UDP() *net.UDPConn
}
