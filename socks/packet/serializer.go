package packet

import (
	"bufio"

	"github.com/josexy/gsocks5/bufferpool"
	"github.com/josexy/gsocks5/socks/constant"
)

var (
	bufferPool = bufferpool.NewBufferPool(func() *[]byte {
		buf := make([]byte, constant.MaxBufferSize)
		return &buf
	})
	udpBufferPool = bufferpool.NewBufferPool(func() *[]byte {
		buf := make([]byte, constant.MaxUdpBufferSize)
		return &buf
	})
)

type Serializer interface {
	String() string
	Release()
	Serialize(buf []byte) []byte
	Revert(data []byte)
}

func GetBuffer(isUdpBuffer bool) *[]byte {
	if isUdpBuffer {
		return udpBufferPool.Get()
	}
	return bufferPool.Get()
}

func ReleaseBuffer(buf *[]byte, isUdpBuffer bool) {
	if isUdpBuffer {
		udpBufferPool.Put(buf)
	}
	bufferPool.Put(buf)
}

func SerializeDirectFrom[T Serializer](buffer []byte) (v T, err error) {
	res := sFactory.New(v.String())
	if res == nil {
		return v, constant.ErrSerializeFailure
	}
	res.Revert(buffer[:])
	return res.(T), nil
}

func SerializeFrom[T Serializer](rw *bufio.ReadWriter) (v T, err error) {
	res := sFactory.New(v.String())
	if res == nil {
		return v, constant.ErrSerializeFailure
	}
	var buffer *[]byte
	if v.String() == StrSocksUDPPacket {
		buffer = udpBufferPool.Get()
		defer udpBufferPool.Put(buffer)
	} else {
		buffer = bufferPool.Get()
		defer bufferPool.Put(buffer)
	}
	n, e := rw.Read(*buffer)
	if e != nil {
		res.Release()
		return v, e
	}
	res.Revert((*buffer)[:n])
	return res.(T), nil
}

func SerializeDirectTo(buffer []byte, v Serializer) (n int, err error) {
	data := v.Serialize(buffer)
	return len(data), nil
}

func SerializeTo(rw *bufio.ReadWriter, v Serializer) (n int, err error) {
	var buffer *[]byte
	if v.String() == StrSocksUDPPacket {
		buffer = udpBufferPool.Get()
		defer udpBufferPool.Put(buffer)
	} else {
		buffer = bufferPool.Get()
		defer bufferPool.Put(buffer)
	}
	n, err = rw.Write(v.Serialize(*buffer))
	_ = rw.Flush()
	return
}
