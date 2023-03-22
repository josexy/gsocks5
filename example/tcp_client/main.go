package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/josexy/gsocks5/socks"
	"github.com/josexy/gsocks5/util"
)

func main() {
	proxyCli := socks.NewSocks5Client("127.0.0.1:10086")
	proxyCli.SetSocksAuth("test", "12345678")
	defer proxyCli.Close()

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return proxyCli.Dial(ctx, addr)
		},
	}
	cli := http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}
	resp, err := cli.Get("http://www.baidu.com")
	if err != nil {
		util.Logger.ErrorBy(err)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}
