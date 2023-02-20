package cgobytepool

/*
#include "bytepool.h"
*/
import "C"

import (
	"runtime"
	"runtime/cgo"
	"sync/atomic"
	"unsafe"
)

type Pool interface {
	Get(int) unsafe.Pointer
	Put(unsafe.Pointer, int)
}

const (
	alignmentSize int = 256
)

func defaultAlign(n int) int {
	return ((n + alignmentSize) >> 3) << 3
}

//export bytepool_get
func bytepool_get(ctx unsafe.Pointer, size C.size_t) unsafe.Pointer {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := int(size)
	return p.Get(defaultAlign(n))
}

//export bytepool_put
func bytepool_put(ctx unsafe.Pointer, data unsafe.Pointer, size C.size_t) {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := defaultAlign(int(size))
	p.Put(data, n)
}

//export bytepool_free
func bytepool_free(ctx unsafe.Pointer) {
	h := *(*cgo.Handle)(ctx)
	h.Delete()
}

var (
	_ Pool = (*defaultPool)(nil)
)

type defaultPool struct {
	pools []*byteslicepool
	bytes int64
}

func (p *defaultPool) find(size int) (*byteslicepool, bool) {
	for _, pp := range p.pools {
		if size <= pp.bufSize {
			return pp, true
		}
	}
	return nil, false
}

func (p *defaultPool) Get(n int) unsafe.Pointer {
	if pp, ok := p.find(n); ok {
		return pp.Get()
	}
	atomic.AddInt64(&p.bytes, int64(n))
	return unsafe.Pointer(C.malloc(C.size_t(n))) // fallback
}

func (p *defaultPool) Put(b unsafe.Pointer, n int) {
	if pp, ok := p.find(n); ok {
		pp.Put(b, n)
		return
	}
	C.free(b) // fallback
	atomic.AddInt64(&p.bytes, -1*int64(n))
}

func (p *defaultPool) AllocBytes() int64 {
	total := int64(0)
	for _, pp := range p.pools {
		total += pp.AllocBytes()
	}
	total += atomic.LoadInt64(&p.bytes)
	return total
}

func (p *defaultPool) Close() {
	for _, pp := range p.pools {
		pp.Close()
	}
}

func NewDefaultPool() *defaultPool {
	poolSize := 1000
	pools := make([]*byteslicepool, 7)
	for i := 0; i < 7; i += 1 {
		pools[i] = newByteSlicePool(poolSize, defaultAlign(4096*(i+1)))
	}
	p := &defaultPool{pools, 0}
	runtime.SetFinalizer(p, func(me *defaultPool) {
		me.Close()
	})
	return p
}

type byteslicepool struct {
	pool    chan unsafe.Pointer
	bufSize int
	bytes   int64
}

func (p *byteslicepool) Get() unsafe.Pointer {
	select {
	case buf := <-p.pool:
		// reuse
		return buf
	default:
		// new
		atomic.AddInt64(&p.bytes, int64(p.bufSize))
		return unsafe.Pointer(C.malloc(C.size_t(p.bufSize)))
	}
}

func (p *byteslicepool) Put(data unsafe.Pointer, size int) {
	if size < p.bufSize {
		C.free(data)
		atomic.AddInt64(&p.bytes, -1*int64(p.bufSize))
		return // discard
	}
	select {
	case p.pool <- data:
		// ok
	default:
		// release
		C.free(data)
		atomic.AddInt64(&p.bytes, -1*int64(p.bufSize))
	}
}

func (p *byteslicepool) AllocBytes() int64 {
	return atomic.LoadInt64(&p.bytes)
}

func (p *byteslicepool) Close() {
	close(p.pool)
	for data := range p.pool {
		C.free(data)
		atomic.AddInt64(&p.bytes, -1*int64(p.bufSize))
	}
}

func newByteSlicePool(poolSize, bufSize int) *byteslicepool {
	return &byteslicepool{
		pool:    make(chan unsafe.Pointer, poolSize),
		bufSize: bufSize,
		bytes:   0,
	}
}
