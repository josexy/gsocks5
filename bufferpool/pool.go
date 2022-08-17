package bufferpool

import (
	"sync"
)

type BufferPool[T any] struct {
	pool sync.Pool
}

func NewBufferPool[T any](newFunc func() T) *BufferPool[T] {
	bp := &BufferPool[T]{
		pool: sync.Pool{},
	}
	bp.pool.New = func() any {
		return newFunc()
	}
	return bp
}

func (bp *BufferPool[T]) Get() T {
	return bp.pool.Get().(T)
}

func (bp *BufferPool[T]) Put(data T) {
	bp.pool.Put(data)
}
