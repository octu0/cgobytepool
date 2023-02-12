package cgobytepool

/*
#include <stdlib.h>

extern void *bytepool_get(void *context, size_t size);
extern void bytepool_put(void *context, void *data, size_t size);
extern void bytepool_free(void *context);

static void benchmarkwrite(unsigned char *p, int size) {
  for(int i = 0; i < size; i += 1) {
    p[i] = 123;
  }
}

static void benchmark_handle(void *ctx, int N) {
  for(int i = 0; i < N; i += 1) {
    unsigned char *p = (unsigned char*)bytepool_get(ctx, 100);
    benchmarkwrite(p, 100);
    bytepool_put(ctx, p, 100);
  }
  bytepool_free(ctx);
}

static void benchmark_malloc(int N) {
  for(int i = 0; i < N; i += 1) {
    unsigned char *p = (unsigned char*) malloc(100);
    benchmarkwrite(p, 100);
    free(p);
  }
}
*/
import "C"

import (
	"runtime/cgo"
	"sync"
	"unsafe"
)

type Pool interface {
	Get(int) []byte
	Put([]byte)
}

//export bytepool_get
func bytepool_get(ctx unsafe.Pointer, size C.size_t) unsafe.Pointer {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := int(size)
	data := p.Get(n)
	return C.CBytes(data)
}

//export bytepool_put
func bytepool_put(ctx unsafe.Pointer, data unsafe.Pointer, size C.size_t) {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	b := C.GoBytes(data, C.int(size))
	p.Put(b)
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
	pool *sync.Pool
}

func (p *defaultPool) Get(n int) []byte {
	b := p.pool.Get().([]byte)
	if n < cap(b) {
		b = make([]byte, 0, n)
	}
	return b[:n]
}

func (p *defaultPool) Put(b []byte) {
	p.pool.Put(b)
}

func NewDefaultPool() *defaultPool {
	return &defaultPool{
		pool: &sync.Pool{
			New: func() any {
				return make([]byte, 0, 32*1024)
			},
		},
	}
}

func benchmarkHandle(N int) {
	p := NewDefaultPool()
	h := cgo.NewHandle(p)
	C.benchmark_handle(unsafe.Pointer(&h), C.int(N))
}

func benchmarkMalloc(N int) {
	C.benchmark_malloc(C.int(N))
}

func benchmarkHandleAndGo(N int) {
	p := NewDefaultPool()
	h := cgo.NewHandle(p)

	half := N / 2
	benchmarkGo(p, half)
	C.benchmark_handle(unsafe.Pointer(&h), C.int(half))
}

func benchmarkwrite(d []byte, size int) {
	for i := 0; i < size; i += 1 {
		d[i] = 123
	}
}

func benchmarkGo(p Pool, N int) {
	for i := 0; i < N; i += 1 {
		d := p.Get(100)
		benchmarkwrite(d, 100)
		p.Put(d)
	}
}
