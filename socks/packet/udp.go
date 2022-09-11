package packet

import (
	"encoding/binary"
	"net"

	"github.com/josexy/gsocks5/socks/constant"
)

type SocksUDPPacket struct {
	AType   constant.Socks5AddressType
	DstAddr string
	DstPort int
	UDPData []byte
}

func (s *SocksUDPPacket) String() string { return StrSocksUDPPacket }

func (s *SocksUDPPacket) Release() { sFactory.Release(s) }

func (s SocksUDPPacket) Serialize(buf []byte) []byte {
	buf[0] = 0x00
	buf[1] = 0x00
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
	vl = index + vl + 2
	vl += copy(buf[vl:], s.UDPData)
	return buf[:vl]
}

func (s *SocksUDPPacket) Revert(data []byte) {
	n := len(data)
	s.AType = data[3]
	var vl int
	switch s.AType {
	case constant.IPv4:
		vl = 4
		s.DstAddr = net.IP(data[4:8]).String()
	case constant.IPv6:
		vl = 16
		s.DstAddr = net.IP(data[4:20]).String()
	case constant.DomainName:
		hostLen := int(data[4]) // domain length
		vl = hostLen + 1
		s.DstAddr = string(data[5 : 5+hostLen])
	}
	s.DstPort = int(binary.BigEndian.Uint16(data[4+vl : 6+vl]))
	size := n - 6 - vl
	s.UDPData = make([]byte, size)
	copy(s.UDPData, data[6+vl:])
}
