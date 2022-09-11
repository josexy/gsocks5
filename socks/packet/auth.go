package packet

import (
	"github.com/josexy/gsocks5/socks/constant"
)

type SocksAuthRequest struct {
	Version  byte
	Username string
	Password string
}

func (s *SocksAuthRequest) String() string { return StrSocksAuthRequest }

func (s *SocksAuthRequest) Release() { sFactory.Release(s) }

func (s *SocksAuthRequest) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version01
	buf[1] = byte(len(s.Username))
	n := copy(buf[2:], s.Username)
	buf[2+n] = byte(len(s.Password))
	n2 := copy(buf[3+n:], s.Password)
	return buf[:3+n+n2]
}

func (s *SocksAuthRequest) Revert(data []byte) {
	s.Version = data[0]
	ul := int(data[1])
	s.Username = string(data[2 : 2+ul])
	pl := int(data[2+ul])
	s.Password = string(data[len(data)-pl:])
}

type SocksAuthResponse struct {
	Version byte
	Status  byte
}

func (s *SocksAuthResponse) String() string { return StrSocksAuthResponse }

func (s *SocksAuthResponse) Release() { sFactory.Release(s) }

func (s SocksAuthResponse) Serialize(buf []byte) []byte {
	buf[0] = constant.Socks5Version01
	buf[1] = s.Status
	return buf[:2]
}

func (s *SocksAuthResponse) Revert(data []byte) {
	s.Version = data[0]
	s.Status = data[1]
}
