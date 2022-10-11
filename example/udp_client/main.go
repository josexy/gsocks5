package main

import (
	"context"
	"fmt"
	"time"

	"github.com/josexy/gsocks5/socks"
	"github.com/josexy/logx"
)

func main() {
	proxyCli := socks.NewSocks5Client("127.0.0.1:10086")
	proxyCli.SetSocksAuth("test", "12345678")
	defer proxyCli.Close()

	conn, err := proxyCli.DialUDP(context.Background(), "127.0.0.1:2003")
	if err != nil {
		logx.ErrorBy(err)
		return
	}
	done := make(chan error)
	go func() {
		buf := make([]byte, 65535)
		for {
			conn.SetReadDeadline(time.Now().Add(time.Second * 2))
			n, err := conn.Read(buf)
			if err != nil {
				done <- err
				return
			}
			fmt.Println(string(buf[:n]))
		}
	}()
	for i := 0; i < 10; i++ {
		conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
		_, err = conn.Write([]byte("hello server " + time.Now().String()))
		if err != nil {
			logx.ErrorBy(err)
			return
		}
	}
	err = <-done
	logx.ErrorBy(err)
}
