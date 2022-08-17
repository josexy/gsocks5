package server

import (
	"bufio"
	"context"
	"io"
	"net"

	"github.com/josexy/gsocks5/logx"
	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/internal/connection"
	"github.com/josexy/gsocks5/socks/internal/packet"
	"github.com/josexy/gsocks5/udpserver"
)

func (s *Socks5Server) handleCmdUdpAssociate(rw *bufio.ReadWriter, target string, src net.Conn) error {
	// [UDP Client <-> Socks Client] <-> [Socks Server] <-> UDP Server
	// 连接到目标UDP Server
	targetConn, err := connection.DialUDP(target)
	if err != nil {
		return err
	}
	defer targetConn.Close()

	var svr *udpserver.UdpServer
	svr, err = udpserver.NewUdpServer("", nil)
	if err != nil {
		return err
	}

	bindAddr := svr.LocalAddr()

	s.server.Opts.Logger.Info("[udp] local: [%s] <-> [%s] <-> remote: [%s]",
		logx.Green(bindAddr.String()),
		logx.Yellow(targetConn.LocalAddr().String()),
		logx.Yellow(target))

	packet.SerializeTo(rw, &packet.SocksResponse{
		ReplayCode: constant.Succeed,
		BindAddr:   bindAddr.IP.String(),
		BindPort:   bindAddr.Port,
	})

	defer svr.Close()

	srcAddrChan := make(chan net.Addr, 5)
	errChan := make(chan error, 2)
	tcpDoneChan := make(chan error)

	// Socks Client <- [Socks Server] <- UDP Server
	go func() {
		bufferRead := packet.GetBuffer(true)
		bufferWrite := packet.GetBuffer(true)
		defer packet.ReleaseBuffer(bufferRead, true)
		defer packet.ReleaseBuffer(bufferWrite, true)

		for {
			n, targetAddr, err := targetConn.ReadFromUDP(*bufferRead)
			if err != nil {
				errChan <- err
				return
			}

			srcAddr := <-srcAddrChan

			var atype constant.Socks5AddressType
			if targetAddr.IP.Equal(targetAddr.IP.To4()) {
				atype = constant.IPv4
			} else {
				atype = constant.IPv6
			}
			sz, _ := packet.SerializeDirectTo(*bufferWrite, &packet.SocksUDPPacket{
				AType:   atype,
				DstAddr: targetAddr.IP.String(),
				DstPort: targetAddr.Port,
				UDPData: (*bufferRead)[:n],
			})
			svr.Conn.WriteTo((*bufferWrite)[:sz], srcAddr)
		}
	}()

	// Socks Client -> [Socks Server] -> UDP Server
	svr.Handler = udpserver.UdpHandlerFunc(func(ctx context.Context, lnConn *net.UDPConn) {
		buffer := packet.GetBuffer(true)
		defer packet.ReleaseBuffer(buffer, true)

		// 读取Socks Client发送的封包数据，并转发给UDP Server
		n, srcAddr, err := lnConn.ReadFrom(*buffer)
		if err != nil {
			errChan <- err
			return
		}

		select {
		case srcAddrChan <- srcAddr:
		default:
		}
		res, err := packet.SerializeDirectFrom[*packet.SocksUDPPacket]((*buffer)[:n])
		if err != nil {
			errChan <- err
			return
		}
		res.Release()

		// 向目标UDP Server发送UDP原始数据报文
		targetConn.Write(res.UDPData)
	})

	go svr.Serve()

	go func() {
		// 等待TCP连接关闭
		buf := make([]byte, 1)
		for {
			_, err = src.Read(buf)
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				tcpDoneChan <- err
				return
			}
		}
	}()

	select {
	case err = <-errChan:
		return err
	case err = <-tcpDoneChan:
		return err
	}
}
