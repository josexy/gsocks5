package constant

const (
	MaxBufferSize    = 515
	MaxUdpBufferSize = 1 << 13
)

const (
	Socks5Version05 = 0x05
	Socks5Version01 = 0x01
)

type Socks5Method = byte

const (
	MethodNoAuthRequired   = 0x00
	MethodGSSAPI           = 0x01
	MethodUsernamePassword = 0x02
	MethodIANAAssigned     = 0x03
	MethodReserved         = 0x80
	MethodNotAcceptable    = 0xFF
)

type Socks5Cmd = byte

const (
	Connect Socks5Cmd = iota + 1
	Bind
	UDP
)

type Socks5AddressType = byte

const (
	IPv4       = 0x01
	DomainName = 0x03
	IPv6       = 0x04
)

type Socks5ReplyCode = byte

const (
	Succeed Socks5ReplyCode = iota
	GeneralSocksServerFailure
	ConnectionNotAllowedByRuleset
	NetworkUnreachable
	HostUnreachable
	ConnectionRefused
	TTLExpired
	CommandNotSupported
	AddressTypeNotSupported
	Unassigned
)
