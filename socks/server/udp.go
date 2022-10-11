package server

import (
	"bufio"
	"io"
	"net"

	"github.com/fatih/color"
	"github.com/josexy/gsocks5/socks/connection"
	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/packet"
	"github.com/josexy/logx"
)

func (s *Socks5Server) serveUDP(conn *net.UDPConn) error {
	buffer := packet.GetBuffer(true)
	defer packet.ReleaseBuffer(buffer, true)

	// Socks Client -> [Socks Server] -> UDP Server
	// 读取Socks Client发送的封包数据，并转发给UDP Server
	n, srcAddr, err := conn.ReadFrom(*buffer)
	if err != nil {
		return err
	}
	res, err := packet.SerializeDirectFrom[*packet.SocksUDPPacket]((*buffer)[:n])
	if err != nil {
		return err
	}
	defer res.Release()

	targetConn := s.natM.Get(srcAddr.String())
	if targetConn == nil {
		var target string
		var ok bool
		if target, ok = <-s.targetAddrChan; !ok {
			return nil
		}
		// 连接到目标UDP Server
		targetConn, err = connection.DialUDP(target)
		if err != nil {
			return err
		}
		// Socks Client <- [Socks Server] <- UDP Server
		s.natM.Add(srcAddr, conn, targetConn)
	}

	// 向目标UDP Server发送UDP原始数据报文
	targetConn.Write(res.UDPData)
	return nil
}

func (s *Socks5Server) handleCmdUdpAssociate(rw *bufio.ReadWriter, target string, src net.Conn) error {
	bindAddr := s.udpServer.LocalAddr()

	logx.Info("[udp] local: [%s] <-> remote: [%s]",
		color.GreenString(bindAddr.String()),
		color.YellowString(target))

	packet.SerializeTo(rw, &packet.SocksResponse{
		ReplayCode: constant.Succeed,
		BindAddr:   bindAddr.IP.String(),
		BindPort:   bindAddr.Port,
	})

	s.targetAddrChan <- target

	tcpDoneChan := make(chan error)
	go func() {
		// 等待TCP连接关闭
		buf := make([]byte, 1)
		for {
			_, err := src.Read(buf)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				tcpDoneChan <- err
				return
			}
		}
	}()

	return <-tcpDoneChan
}
