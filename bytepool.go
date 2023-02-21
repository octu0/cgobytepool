package cgobytepool

/*
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"runtime/cgo"
	"sort"
	"sync"
	"sync/atomic"
	"unsafe"
)

type PoolStats struct {
	Allocs []struct {
		ID   int
		Size int64
	}
	Fallback struct {
		ID   int
		Size int64
	}
}

type Pool interface {
	Get(int) unsafe.Pointer
	Put(unsafe.Pointer, int)
	Close()

	Stats() PoolStats
}

func HandlePoolGet(ctx unsafe.Pointer, size int) unsafe.Pointer {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	return p.Get(size)
}

func HandlePoolPut(ctx unsafe.Pointer, data unsafe.Pointer, size int) {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	p.Put(data, size)
}

func HandlePoolFree(ctx unsafe.Pointer) {
	h := *(*cgo.Handle)(ctx)
	h.Delete()
}

func CgoHandle(p Pool) cgo.Handle {
	return cgo.NewHandle(p)
}

type MemoryAligmentFunc func(int) int

type WithPoolFunc func(MemoryAligmentFunc) *cmallocPool

func WithPoolSize(poolSize, bufferSize int) WithPoolFunc {
	return func(fn MemoryAligmentFunc) *cmallocPool {
		return newCMallocPool(poolSize, fn(bufferSize))
	}
}

const (
	defaultMemoryAlignmentSize int = 256
)

var (
	DefaultMemoryAlignmentFunc MemoryAligmentFunc = func(n int) int {
		return ((n + defaultMemoryAlignmentSize) >> 3) << 3
	}
)

var (
	_ Pool = (*CgoBytePool)(nil)
)

type CgoBytePool struct {
	pools     []*cmallocPool
	bytes     int64
	alignFunc MemoryAligmentFunc
	fallbacks *sync.Map // map[uintptr]unsafe.Pointer
}

func (p *CgoBytePool) find(size int) (*cmallocPool, bool) {
	for _, pp := range p.pools {
		if size <= pp.bufSize {
			return pp, true
		}
	}
	return nil, false
}

func (p *CgoBytePool) Get(size int) unsafe.Pointer {
	n := p.alignFunc(size)
	if pp, ok := p.find(n); ok {
		return pp.Get()
	}
	return p.fallbackGet(n)
}

func (p *CgoBytePool) fallbackGet(n int) unsafe.Pointer {
	atomic.AddInt64(&p.bytes, int64(n))
	ptr := unsafe.Pointer(C.malloc(C.size_t(n)))
	p.fallbacks.Store(uintptr(ptr), ptr)
	return ptr
}

func (p *CgoBytePool) Put(b unsafe.Pointer, size int) {
	n := p.alignFunc(size)
	if pp, ok := p.find(n); ok {
		pp.Put(b, n)
		return
	}
	p.fallbackPut(b, n)
}

func (p *CgoBytePool) fallbackPut(b unsafe.Pointer, n int) {
	v, ok := p.fallbacks.LoadAndDelete(uintptr(b))
	if ok {
		ptr := v.(unsafe.Pointer)
		C.free(ptr)
		atomic.AddInt64(&p.bytes, -1*int64(n))
	}
}

func (p *CgoBytePool) Stats() PoolStats {
	ps := PoolStats{
		Allocs: make([]struct {
			ID   int
			Size int64
		}, len(p.pools)),
	}

	for i, pp := range p.pools {
		ps.Allocs[i].ID = i
		ps.Allocs[i].Size = pp.AllocBytes()
	}
	ps.Fallback.ID = 0
	ps.Fallback.Size = p.AllocBytes()
	return ps
}

func (p *CgoBytePool) AllocBytes() int64 {
	return atomic.LoadInt64(&p.bytes)
}

func (p *CgoBytePool) TotalAllocBytes() int64 {
	total := int64(0)
	for _, pp := range p.pools {
		total += pp.AllocBytes()
	}
	total += p.AllocBytes()
	return total
}

func (p *CgoBytePool) Close() {
	runtime.SetFinalizer(p, nil) // clear finalizer
	for _, pp := range p.pools {
		pp.Close()
	}
}

func finalizeDefaultPool(p *CgoBytePool) {
	p.Close()
}

func NewPool(alignFunc MemoryAligmentFunc, poolFuncs ...WithPoolFunc) *CgoBytePool {
	if alignFunc == nil {
		alignFunc = DefaultMemoryAlignmentFunc
	}

	pools := make([]*cmallocPool, len(poolFuncs))
	for i, fn := range poolFuncs {
		pools[i] = fn(alignFunc)
	}
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].bufSize < pools[j].bufSize // order bufSize asc
	})

	p := &CgoBytePool{pools, 0, alignFunc, new(sync.Map)}
	runtime.SetFinalizer(p, finalizeDefaultPool)
	return p
}

type cmallocPool struct {
	pool    chan unsafe.Pointer
	bufSize int
	bytes   int64
}

func (p *cmallocPool) Get() unsafe.Pointer {
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

func (p *cmallocPool) Put(data unsafe.Pointer, size int) {
	if size < p.bufSize {
		C.free(data)
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

func (p *cmallocPool) AllocBytes() int64 {
	return atomic.LoadInt64(&p.bytes)
}

func (p *cmallocPool) Close() {
	close(p.pool)
	for data := range p.pool {
		C.free(data)
		atomic.AddInt64(&p.bytes, -1*int64(p.bufSize))
	}
}

func newCMallocPool(poolSize, bufSize int) *cmallocPool {
	return &cmallocPool{
		pool:    make(chan unsafe.Pointer, poolSize),
		bufSize: bufSize,
		bytes:   0,
	}
}
