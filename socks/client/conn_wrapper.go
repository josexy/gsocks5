package client

import (
	"bufio"
	"net"

	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/packet"
)

type tcpConnWrapper struct {
	net.Conn
	remoteAddr net.Addr // target address
}

func newTcpConnWrapper(conn net.Conn, target string) (*tcpConnWrapper, error) {
	addr, err := net.ResolveTCPAddr("tcp", target)
	if err != nil {
		return nil, err
	}
	return &tcpConnWrapper{
		Conn:       conn,
		remoteAddr: addr,
	}, nil
}

func (c *tcpConnWrapper) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *tcpConnWrapper) TCP() *net.TCPConn {
	return c.Conn.(*net.TCPConn)
}

func (c *tcpConnWrapper) UDP() *net.UDPConn {
	return nil
}

type udpConnWrapper struct {
	*net.UDPConn
	rw         *bufio.ReadWriter
	remoteAddr net.Addr // target address
}

func newUdpConnWrapper(conn *net.UDPConn, target string) (*udpConnWrapper, error) {
	addr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}
	return &udpConnWrapper{
		UDPConn:    conn,
		rw:         bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		remoteAddr: addr,
	}, nil
}

func (c *udpConnWrapper) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *udpConnWrapper) TCP() *net.TCPConn {
	return nil
}

func (c *udpConnWrapper) UDP() *net.UDPConn {
	return c.UDPConn
}

func (c *udpConnWrapper) Read(b []byte) (int, error) {
	res, err := packet.SerializeFrom[*packet.SocksUDPPacket](c.rw)
	if err != nil {
		return 0, err
	}
	defer res.Release()
	n := copy(b, res.UDPData)
	return n, nil
}

func (c *udpConnWrapper) Write(b []byte) (int, error) {
	addr := c.remoteAddr.(*net.UDPAddr)
	var atype constant.Socks5AddressType
	if addr.IP.Equal(addr.IP.To4()) {
		atype = constant.IPv4
	} else {
		atype = constant.IPv6
	}
	return packet.SerializeTo(c.rw, &packet.SocksUDPPacket{
		AType:   atype,
		DstAddr: addr.IP.String(),
		DstPort: addr.Port,
		UDPData: b,
	})
}
