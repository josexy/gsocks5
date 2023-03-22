package server

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"time"

	"github.com/josexy/gsocks5/config"
	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/packet"
	"github.com/josexy/gsocks5/socks/sc"
	"github.com/josexy/gsocks5/tcpserver"
	"github.com/josexy/gsocks5/udpserver"
	"github.com/josexy/gsocks5/util"
)

type Socks5Server struct {
	server         *tcpserver.TcpServer
	udpServer      *udpserver.UdpServer
	targetAddrChan chan string
	natM           *sc.UdpNATMap
}

func NewSocks5Server(addr string) (svr *Socks5Server) {
	svr = &Socks5Server{
		targetAddrChan: make(chan string, 128),
		natM:           sc.NewUdpNATMap(time.Second * 20),
	}
	svr.server = tcpserver.NewTcpServer(addr, svr)
	svr.udpServer, _ = udpserver.NewUdpServer(addr, svr)
	return
}

func (s *Socks5Server) Start() error {
	go s.udpServer.Serve()
	return s.server.ListenAndServe()
}

func (s *Socks5Server) Close() error {
	s.udpServer.Close()
	return s.server.Close()
}

func (s *Socks5Server) ServeTCP(ctx context.Context, conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	if err := s.handleNegotiate(rw); err != nil {
		util.Logger.ErrorBy(err)
		return
	}
	if err := s.handleRequest(rw, conn); err != nil {
		util.Logger.ErrorBy(err)
		return
	}
}

func (s *Socks5Server) ServeUDP(ctx context.Context, conn *net.UDPConn) {
	if err := s.serveUDP(conn); err != nil {
		util.Logger.ErrorBy(err)
	}
}

func (s *Socks5Server) chooseMethod(clientMethod, serverMethod []constant.Socks5Method) constant.Socks5Method {
	var method constant.Socks5Method
	if len(serverMethod) == 0 {
		method = constant.MethodNoAuthRequired
	} else {
		method = serverMethod[0]
	}
	for _, m := range clientMethod {
		if m == method {
			return method
		}
	}
	if len(clientMethod) == 0 {
		return constant.MethodNoAuthRequired
	}
	return clientMethod[0]
}

func (s *Socks5Server) handleNegotiate(rw *bufio.ReadWriter) error {
	res, err := packet.SerializeFrom[*packet.SocksNegotiateRequest](rw)
	if err != nil {
		return err
	}
	defer res.Release()
	if res.Version != constant.Socks5Version05 {
		return constant.ErrVersion5Invalid
	}
	if res.NMethods < 0 {
		return constant.ErrUnsupportedMethod
	}
	method := s.chooseMethod(res.Methods, config.Cfg.SocksMethod)
	packet.SerializeTo(rw, &packet.SocksNegotiateResponse{
		Method: method,
	})
	if method == constant.MethodUsernamePassword {
		return s.handleAuth(rw)
	}
	return nil
}

func (s *Socks5Server) handleAuth(rw *bufio.ReadWriter) error {
	res, err := packet.SerializeFrom[*packet.SocksAuthRequest](rw)
	if err != nil {
		return err
	}
	defer res.Release()
	if res.Version != constant.Socks5Version01 {
		return constant.ErrVersion1Invalid
	}

	for _, auth := range config.Cfg.Auth {
		if auth.Auth(res.Username, res.Password) {
			packet.SerializeTo(rw, &packet.SocksAuthResponse{})
			return nil
		}
	}
	packet.SerializeTo(rw, &packet.SocksAuthResponse{
		Status: constant.GeneralSocksServerFailure,
	})
	return constant.ErrAuthFailure
}

func (s *Socks5Server) handleRequest(rw *bufio.ReadWriter, src net.Conn) error {
	res, err := packet.SerializeFrom[*packet.SocksRequest](rw)
	if err != nil {
		return err
	}
	if res == nil {
		return constant.ErrSerializeFailure
	}
	defer res.Release()
	if res.Version != constant.Socks5Version05 {
		return constant.ErrVersion5Invalid
	}
	switch res.AType {
	case constant.IPv4, constant.IPv6, constant.DomainName:
	default:
		return constant.ErrUnsupportedReqAType
	}

	target := net.JoinHostPort(res.DstAddr, strconv.Itoa(res.DstPort))
	switch res.Cmd {
	case constant.Connect:
		if err = s.handleCmdConnect(rw, target, src); err != nil {
			return err
		}
	case constant.UDP:
		if err = s.handleCmdUdpAssociate(rw, target, src); err != nil {
			return err
		}
	//case constant.Bind:
	default:
		packet.SerializeTo(rw, &packet.SocksResponse{ReplayCode: constant.CommandNotSupported})
		return constant.ErrUnsupportedReqCmd
	}

	return nil
}
