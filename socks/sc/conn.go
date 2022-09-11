package sc

import (
	"net"
	"strconv"

	"github.com/josexy/gsocks5/socks/constant"
)

type Conn interface {
	net.Conn
	TCP() *net.TCPConn
	UDP() *net.UDPConn
}

type Address []byte

func ParseAddress(addr string) Address {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil
	}
	var buf Address
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buf = make([]byte, 1+net.IPv4len+2)
			buf[0] = constant.IPv4
			copy(buf[1:], ip4)
		} else {
			buf = make([]byte, 1+net.IPv6len+2)
			buf[0] = constant.IPv6
			copy(buf[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return nil
		}
		buf = make([]byte, 1+1+len(host)+2)
		buf[0] = constant.DomainName
		buf[1] = byte(len(host))
		copy(buf[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil
	}

	buf[len(buf)-2], buf[len(buf)-1] = byte(portnum>>8), byte(portnum)
	return buf
}

func (a Address) String() string {
	var host, port string

	switch a[0] {
	case constant.DomainName:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case constant.IPv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case constant.IPv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}
