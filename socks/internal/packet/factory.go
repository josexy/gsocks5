package packet

import (
	"sync"

	"github.com/josexy/gsocks5/bufferpool"
)

const (
	StrSocksNegotiateRequest  = "SocksNegotiateRequest"
	StrSocksNegotiateResponse = "SocksNegotiateResponse"
	StrSocksAuthRequest       = "SocksAuthRequest"
	StrSocksAuthResponse      = "SocksAuthResponse"
	StrSocksRequest           = "SocksRequest"
	StrSocksResponse          = "SocksResponse"
	StrSocksUDPPacket         = "SocksUDPPacket"
)

var sFactory = newSerializerFactory()

type SerializerFactory interface {
	New(string) Serializer
	Release(Serializer)
}

type simpleSerializerFactory struct {
	record map[string]*bufferpool.BufferPool[Serializer]
	mu     sync.RWMutex
}

func newSerializerFactory() *simpleSerializerFactory {
	return &simpleSerializerFactory{
		record: map[string]*bufferpool.BufferPool[Serializer]{
			StrSocksNegotiateRequest:  bufferpool.NewBufferPool(func() Serializer { return new(SocksNegotiateRequest) }),
			StrSocksNegotiateResponse: bufferpool.NewBufferPool(func() Serializer { return new(SocksNegotiateResponse) }),
			StrSocksAuthRequest:       bufferpool.NewBufferPool(func() Serializer { return new(SocksAuthRequest) }),
			StrSocksAuthResponse:      bufferpool.NewBufferPool(func() Serializer { return new(SocksAuthResponse) }),
			StrSocksRequest:           bufferpool.NewBufferPool(func() Serializer { return new(SocksRequest) }),
			StrSocksResponse:          bufferpool.NewBufferPool(func() Serializer { return new(SocksResponse) }),
			StrSocksUDPPacket:         bufferpool.NewBufferPool(func() Serializer { return new(SocksUDPPacket) }),
		},
	}
}

func (sf *simpleSerializerFactory) New(name string) Serializer {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	if bp, ok := sf.record[name]; ok {
		return bp.Get()
	}
	return nil
}

func (sf *simpleSerializerFactory) Release(obj Serializer) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	if bp, ok := sf.record[obj.String()]; ok {
		bp.Put(obj)
	}
}
