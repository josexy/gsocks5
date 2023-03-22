package main

import (
	"fmt"
	"net"

	"github.com/josexy/gsocks5/util"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":2003")
	if err != nil {
		util.Logger.ErrorBy(err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		util.Logger.ErrorBy(err)
		return
	}
	defer conn.Close()

	util.Logger.Infof("remote addr: %v", conn.RemoteAddr())
	util.Logger.Infof("local addr: %v", conn.LocalAddr())

	buf := make([]byte, 65535)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			util.Logger.ErrorBy(err)
			return
		}
		conn.WriteTo(buf[:n], addr)
		fmt.Println(addr, "->", string(buf[:n]))
	}
}
