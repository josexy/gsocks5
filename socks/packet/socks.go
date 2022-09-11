package packet

import (
	"encoding/binary"
	"net"

	"github.com/josexy/gsocks5/socks/constant"
)

type SocksRequest struct {
	Version byte
	Cmd     constant.Socks5Cmd
	AType   constant.Socks5AddressType
	DstAddr string
	DstPort int
}

func (s *SocksRequest) String() string { return StrSocksRequest }

func (s *SocksRequest) Release() { sFactory.Release(s) }

func (s SocksRequest) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version05
	buf[1] = s.Cmd
	buf[2] = 0x00
	buf[3] = s.AType

	var vl int
	var ip net.IP
	switch s.AType {
	case constant.IPv4:
		vl = 4
		ip = net.ParseIP(s.DstAddr).To4()
	case constant.IPv6:
		vl = 16
		ip = net.ParseIP(s.DstAddr).To16()
	case constant.DomainName:
		vl = len(s.DstAddr) + 1
		ip = []byte(s.DstAddr)
	}

	index := 4
	if s.AType == constant.DomainName {
		buf[index] = byte(len(s.DstAddr))
		index++
	}
	vl = copy(buf[index:], ip)
	binary.BigEndian.PutUint16(buf[index+vl:], uint16(s.DstPort))
	return buf[:index+vl+2]
}

func (s *SocksRequest) Revert(data []byte) {
	n := len(data)
	s.Version = data[0]
	s.Cmd = data[1]
	s.AType = data[3]
	switch s.AType {
	case constant.IPv4:
		s.DstAddr = net.IP(data[4:8]).String()
	case constant.IPv6:
		s.DstAddr = net.IP(data[4:20]).String()
	case constant.DomainName:
		hostLen := int(data[4]) // domain length
		s.DstAddr = string(data[5 : 5+hostLen])
	}
	s.DstPort = int(binary.BigEndian.Uint16(data[n-2:]))
}

type SocksResponse struct {
	Version    byte
	ReplayCode constant.Socks5ReplyCode
	AType      constant.Socks5AddressType
	BindAddr   string
	BindPort   int
}

func (s *SocksResponse) String() string { return StrSocksResponse }

func (s *SocksResponse) Release() { sFactory.Release(s) }

func (s SocksResponse) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version05
	buf[1] = s.ReplayCode
	buf[2] = 0x00

	var vl int
	var atype constant.Socks5AddressType
	ip := net.ParseIP(s.BindAddr)
	if ip.Equal(ip.To4()) {
		// ipv4
		vl = 4
		ip = ip.To4()
		atype = constant.IPv4
	} else {
		// ipv6
		vl = 16
		ip = ip.To16()
		atype = constant.IPv6
	}
	buf[3] = atype

	vl = copy(buf[4:], ip)
	binary.BigEndian.PutUint16(buf[4+vl:], uint16(s.BindPort))
	return buf[:6+vl]
}

func (s *SocksResponse) Revert(data []byte) {
	n := len(data)
	s.Version = data[0]
	s.ReplayCode = data[1]
	s.AType = data[3]
	switch s.AType {
	case constant.IPv4:
		s.BindAddr = net.IP(data[4:8]).String()
	case constant.IPv6:
		s.BindAddr = net.IP(data[4:20]).String()
	case constant.DomainName:
		s.BindAddr = string(data[4 : n-2])
	}
	s.BindPort = int(binary.BigEndian.Uint16(data[n-2:]))
}
