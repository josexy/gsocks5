package packet

import "github.com/josexy/gsocks5/socks/constant"

type SocksNegotiateRequest struct {
	Version  byte
	NMethods int
	Methods  []constant.Socks5Method
}

func (s *SocksNegotiateRequest) String() string { return StrSocksNegotiateRequest }

func (s *SocksNegotiateRequest) Release() { sFactory.Release(s) }

func (s SocksNegotiateRequest) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version05
	buf[1] = byte(s.NMethods)
	copy(buf[2:], s.Methods)
	return buf[:2+s.NMethods]
}

func (s *SocksNegotiateRequest) Revert(data []byte) {
	s.Version = data[0]
	s.NMethods = int(data[1])
	s.Methods = data[2:]
}

type SocksNegotiateResponse struct {
	Version byte
	Method  constant.Socks5Method
}

func (s *SocksNegotiateResponse) String() string { return StrSocksNegotiateResponse }

func (s *SocksNegotiateResponse) Release() { sFactory.Release(s) }

func (s SocksNegotiateResponse) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version05
	buf[1] = s.Method
	return buf[:2]
}

func (s *SocksNegotiateResponse) Revert(data []byte) {
	s.Version = data[0]
	s.Method = data[1]
}
