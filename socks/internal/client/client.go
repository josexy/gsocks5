package client

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/josexy/gsocks5/socks/auth"
	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/internal/connection"
	"github.com/josexy/gsocks5/socks/internal/packet"
	"github.com/josexy/gsocks5/socks/sc"
	"github.com/josexy/gsocks5/util"
)

var defaultSupportMethods = []constant.Socks5Method{
	constant.MethodNoAuthRequired,
	constant.MethodUsernamePassword,
}

type Socks5Client struct {
	Addr string

	conn       net.Conn
	udpConn    net.Conn
	timeout    time.Duration
	dialer     *net.Dialer
	authMethod constant.Socks5Method
	authInfo   auth.Socks5Auth
}

func NewSocks5Client(addr string) *Socks5Client {
	timeout := time.Second * 10
	return &Socks5Client{
		Addr:       addr,
		timeout:    timeout,
		authMethod: constant.MethodNoAuthRequired,
		dialer: &net.Dialer{
			Timeout: timeout,
		},
	}
}

func (c *Socks5Client) SetSocksAuth(username, password string) {
	c.authInfo = auth.NewSocksAuth(username, password)
	c.authMethod = constant.MethodUsernamePassword
}

func (c *Socks5Client) Close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()
	}
	if c.udpConn != nil {
		err = c.udpConn.Close()
	}
	return
}

func (c *Socks5Client) Dial(ctx context.Context, addr string) (sc.Conn, error) {
	_, err := c.handshake(ctx, "tcp", addr, constant.Connect)
	if err != nil {
		return nil, err
	}
	tcw, err := newTcpConnWrapper(c.conn, addr)
	c.conn = tcw
	return tcw, err
}

func (c *Socks5Client) DialUDP(ctx context.Context, addr string) (sc.Conn, error) {
	bindAddr, err := c.handshake(ctx, "tcp", addr, constant.UDP)
	if err != nil {
		return nil, err
	}
	conn, err := connection.DialUDP(bindAddr)
	if err != nil {
		return nil, err
	}
	ucw, err := newUdpConnWrapper(conn, addr)
	c.udpConn = ucw
	return ucw, err
}

func (c *Socks5Client) handshake(ctx context.Context, network, address string, cmd constant.Socks5Cmd) (string, error) {
	conn, err := connection.Dial(ctx, network, c.Addr, c.timeout)
	if err != nil {
		return "", err
	}
	c.conn = conn
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	rw := bufio.NewReadWriter(reader, writer)

	if err = c.negotiate(rw); err != nil {
		_ = conn.Close()
		return "", err
	}

	if err = c.authentication(rw); err != nil {
		_ = conn.Close()
		return "", err
	}
	var bindAddr string
	if bindAddr, err = c.handleRequest(rw, address, cmd); err != nil {
		_ = conn.Close()
		return "", err
	}
	return bindAddr, nil
}

func (c *Socks5Client) negotiate(rw *bufio.ReadWriter) error {
	packet.SerializeTo(rw, &packet.SocksNegotiateRequest{
		NMethods: len(defaultSupportMethods),
		Methods:  defaultSupportMethods,
	})

	res, err := packet.SerializeFrom[*packet.SocksNegotiateResponse](rw)
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

	c.authMethod = res.Method
	return nil
}

func (c *Socks5Client) authentication(rw *bufio.ReadWriter) error {
	if c.authMethod != constant.MethodUsernamePassword {
		return nil
	}

	packet.SerializeTo(rw, &packet.SocksAuthRequest{
		Username: c.authInfo.Username,
		Password: c.authInfo.Password,
	})

	res, err := packet.SerializeFrom[*packet.SocksAuthResponse](rw)
	if err != nil {
		return err
	}
	defer res.Release()
	if res.Version != constant.Socks5Version01 {
		return constant.ErrVersion1Invalid
	}
	if res.Status != 0x00 {
		return constant.ErrAuthFailure
	}
	return nil
}

func (c *Socks5Client) handleRequest(rw *bufio.ReadWriter, target string, cmd constant.Socks5Cmd) (string, error) {
	var host string
	var port int
	hp := strings.Split(target, ":")
	host = hp[0]
	if len(hp) == 1 {
		port = 80
	} else {
		port, _ = strconv.Atoi(hp[1])
	}
	packet.SerializeTo(rw, &packet.SocksRequest{
		Cmd:     cmd,
		AType:   constant.IPv4,
		DstAddr: util.ResolveDomain(host),
		DstPort: port,
	})

	res, err := packet.SerializeFrom[*packet.SocksResponse](rw)
	if err != nil {
		return "", err
	}
	defer res.Release()
	if res.ReplayCode != constant.Succeed {
		return "", constant.ErrRequestFailure
	}
	return net.JoinHostPort(res.BindAddr, strconv.Itoa(res.BindPort)), nil
}
