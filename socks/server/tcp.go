package server

import (
	"bufio"
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/josexy/gsocks5/socks/connection"
	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/packet"
	"github.com/josexy/gsocks5/util"
)

func (s *Socks5Server) handleCmdConnect(rw *bufio.ReadWriter, target string, src net.Conn) error {
	dest, bindAddr, bindPort, err := s.dialTCP(target)
	if err != nil {
		return err
	}
	packet.SerializeTo(rw, &packet.SocksResponse{
		ReplayCode: constant.Succeed,
		BindAddr:   bindAddr,
		BindPort:   bindPort,
	})
	s.forwardData(dest, src)
	return nil
}

func (s *Socks5Server) dialTCP(target string) (conn net.Conn, bindAddr string, bindPort int, err error) {
	conn, err = connection.Dial(context.Background(),
		"tcp",
		target,
		time.Second*10,
	)
	if err != nil {
		return
	}
	if addr, ok := conn.LocalAddr().(*net.TCPAddr); ok {
		bindAddr = addr.IP.String()
		bindPort = addr.Port
	}
	util.Logger.Infof("[tcp] local: [%s] <-> remote: [%s]/[%s]",
		color.GreenString(net.JoinHostPort(bindAddr, strconv.Itoa(bindPort))),
		color.YellowString(target),
		color.RedString(conn.RemoteAddr().String()))
	return
}

func (s *Socks5Server) forwardData(dest, src net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	s.forward(dest, src, &wg)
	wg.Wait()
}

func (s *Socks5Server) forward(dest, src net.Conn, wg *sync.WaitGroup) {
	fn := func(dest, src net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(dest, src)
		_ = dest.Close()
	}
	go fn(dest, src)
	go fn(src, dest)
}
