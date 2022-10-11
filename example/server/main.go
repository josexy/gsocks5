package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/josexy/gsocks5/config"
	"github.com/josexy/gsocks5/socks"
	"github.com/josexy/gsocks5/tcpserver"
	"github.com/josexy/logx"
)

func _main() {
	svr := socks.NewSocks5Server(":10086")
	defer svr.Close()
	svr.Start()
}

var configFile string

func main() {
	flag.StringVar(&configFile, "c", "./config.yaml", "socks5 server config file")
	flag.Parse()

	config.ParseConfig(configFile)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	svr := socks.NewSocks5Server(config.Cfg.ListenAddr)
	logx.Info("start socks server: %s", config.Cfg.ListenAddr)

	done := make(chan struct{})
	go func() {
		if err := svr.Start(); err != nil && err != tcpserver.ErrServerClosed {
			logx.ErrorBy(err)
		}
		done <- struct{}{}
	}()

	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	svr.Close()

	select {
	case <-ctx.Done():
		logx.Warn("socks5 server close timeout")
	case <-done:
		logx.Info("socks5 server closed")
	}
}
