package sc

import (
	"net"
	"sync"
	"time"

	"github.com/josexy/gsocks5/socks/constant"
	"github.com/josexy/gsocks5/socks/packet"
)

type UdpNATMap struct {
	sync.RWMutex
	m       map[string]*net.UDPConn
	timeout time.Duration
}

func NewUdpNATMap(timeout time.Duration) *UdpNATMap {
	return &UdpNATMap{
		m:       make(map[string]*net.UDPConn),
		timeout: timeout,
	}
}

func (m *UdpNATMap) Get(srcAddr string) *net.UDPConn {
	m.RLock()
	defer m.RUnlock()
	return m.m[srcAddr]
}

func (m *UdpNATMap) Set(srcAddr string, targetConn *net.UDPConn) {
	m.Lock()
	defer m.Unlock()

	m.m[srcAddr] = targetConn
}

func (m *UdpNATMap) Del(srcAddr string) *net.UDPConn {
	m.Lock()
	defer m.Unlock()
	if pc, ok := m.m[srcAddr]; ok {
		delete(m.m, srcAddr)
		return pc
	}
	return nil
}

func (m *UdpNATMap) Add(srcAddr net.Addr, dst, src *net.UDPConn) {
	m.Set(srcAddr.String(), src)

	go func() {
		// srcAddr <- dst <- src
		m.forward(srcAddr, dst, src)
		if conn := m.Del(srcAddr.String()); conn != nil {
			conn.Close()
		}
	}()
}

func (m *UdpNATMap) forward(srcAddr net.Addr, dst *net.UDPConn, src *net.UDPConn) error {
	bufferRead := packet.GetBuffer(true)
	bufferWrite := packet.GetBuffer(true)
	defer packet.ReleaseBuffer(bufferRead, true)
	defer packet.ReleaseBuffer(bufferWrite, true)

	for {
		src.SetReadDeadline(time.Now().Add(m.timeout))
		n, targetAddr, err := src.ReadFromUDP(*bufferRead)
		if err != nil {
			return err
		}

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
		dst.WriteTo((*bufferWrite)[:sz], srcAddr)
	}
}
