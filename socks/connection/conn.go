package connection

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

func Dial(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	ctx, cancel := context.WithTimeout(ctx, dialer.Timeout)
	defer cancel()
	con, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return con, nil
}

func DialUDP(address string) (*net.UDPConn, error) {
	rAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	con, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return nil, err
	}
	return con, nil
}

func DialTLS(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout: timeout,
		},
		Config: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	con, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return con, err
}
