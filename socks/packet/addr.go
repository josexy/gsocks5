package packet

import (
	"fmt"
	"net"
	"strconv"

	"github.com/josexy/gsocks5/socks/constant"
)

type AddrSpec struct {
	FQDN     string
	IP       net.IP
	Port     int
	AddrType constant.Socks5AddressType
}

func (a AddrSpec) String() string {
	if len(a.IP) != 0 {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

func (a AddrSpec) Address() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

func ParseAddrSpec(addr string) (as AddrSpec, err error) {
	var host, port string

	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return
	}
	as.Port, err = strconv.Atoi(port)
	if err != nil {
		return
	}

	ip := net.ParseIP(host)
	if ip4 := ip.To4(); ip4 != nil {
		as.AddrType, as.IP = constant.IPv4, ip
	} else if ip6 := ip.To16(); ip6 != nil {
		as.AddrType, as.IP = constant.IPv6, ip
	} else {
		as.AddrType, as.FQDN = constant.DomainName, host
	}
	return
}
