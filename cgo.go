package cgobytepool

/*
#include <stdlib.h>
#include <string.h>

extern void *bytepool_get(void *context, size_t size);
extern void bytepool_put(void *context, void *data, size_t size);
extern void bytepool_free(void *context);

typedef struct foo_t {
  unsigned char *data1;
  unsigned char *data2;
  unsigned char *data3;
  int size1;
  int size2;
  int size3;
} foo_t;

static void test_write_foo(foo_t *foo) {
  foo->data1[0] = 110;
  foo->data1[1000] = 111;
  foo->data1[16383] = 112;
  foo->data2[0] = 120;
  foo->data2[500] = 121;
  foo->data2[4095] = 122;
  foo->data3[0] = 130;
  foo->data3[510] = 131;
  foo->data3[511] = 132;
}

static foo_t *alloc_bytepool(void *ctx) {
  foo_t *foo = (foo_t *) malloc(sizeof(foo_t));
  memset(foo, 0, sizeof(foo_t));

  int size1 = 16 * 1024;
  int size2 = 4 * 1024;
  int size3 = 512;
  foo->data1 = (unsigned char *) bytepool_get(ctx, size1);
  foo->size1 = size1;
  foo->data2 = (unsigned char *) bytepool_get(ctx, size2);
  foo->size2 = size2;
  foo->data3 = (unsigned char *) bytepool_get(ctx, size3);
  foo->size3 = size3;
  test_write_foo(foo);
  return foo;
}

static void free_bytepool(void *ctx, foo_t *foo) {
  if(foo != NULL) {
    bytepool_put(ctx, foo->data1, foo->size1);
    bytepool_put(ctx, foo->data2, foo->size2);
    bytepool_put(ctx, foo->data3, foo->size3);
  }
  free(foo);
}

static foo_t *alloc_malloc() {
  foo_t *foo = (foo_t *) malloc(sizeof(foo_t));
  memset(foo, 0, sizeof(foo_t));

  int size1 = 16 * 1024;
  int size2 = 4 * 1024;
  int size3 = 512;
  foo->data1 = (unsigned char *) malloc(size1);
  foo->size1 = size1;
  foo->data2 = (unsigned char *) malloc(size2);
  foo->size2 = size2;
  foo->data3 = (unsigned char *) malloc(size3);
  foo->size3 = size3;
  test_write_foo(foo);
  return foo;
}

static void free_malloc(foo_t *foo) {
  if(foo != NULL) {
    free(foo->data1);
    free(foo->data2);
    free(foo->data3);
  }
  free(foo);
}
*/
import "C"

import (
	"reflect"
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

func checkCdata1(data []byte) {
	if len(data) != 16384 {
		panic("data1 length 16384")
	}
	if data[0] != 110 {
		panic("data1[0] 110")
	}
	if data[1000] != 111 {
		panic("data1[1000] 111")
	}
	if data[16383] != 112 {
		panic("data1[16383] 112")
	}
}

func checkCdata2(data []byte) {
	if len(data) != 4096 {
		panic("data2 length 4096")
	}
	if data[0] != 120 {
		panic("data2[0] 120")
	}
	if data[500] != 121 {
		panic("data2[500] 121")
	}
	if data[4095] != 122 {
		panic("data2[4095] 122")
	}
}

func checkCdata3(data []byte) {
	if len(data) != 512 {
		panic("data3 length 512")
	}
	if data[0] != 130 {
		panic("data3[0] 130")
	}
	if data[510] != 131 {
		panic("data3[510] 131")
	}
	if data[511] != 132 {
		panic("data3[511] 132")
	}
}

func benchmarkHandle(p Pool) {
	h := cgo.NewHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_bytepool(ctx)))
	defer C.free_bytepool(ctx, foo)

	data1 := C.GoBytes(unsafe.Pointer(foo.data1), C.int(foo.size1))
	data2 := C.GoBytes(unsafe.Pointer(foo.data2), C.int(foo.size2))
	data3 := C.GoBytes(unsafe.Pointer(foo.data3), C.int(foo.size3))
	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkHandleReflect(p Pool) {
	h := cgo.NewHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	fooptr := unsafe.Pointer(C.alloc_bytepool(ctx))
	foo := (*C.foo_t)(fooptr)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = defaultAlign(int(foo.size1))
	s1.Len = int(foo.size1)
	s1.Data = uintptr(unsafe.Pointer(foo.data1))

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = defaultAlign(int(foo.size2))
	s2.Len = int(foo.size2)
	s2.Data = uintptr(unsafe.Pointer(foo.data2))

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = defaultAlign(int(foo.size3))
	s3.Len = int(foo.size3)
	s3.Data = uintptr(unsafe.Pointer(foo.data3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(unsafe.Pointer(foo.data1), defaultAlign(int(foo.size1)))
	p.Put(unsafe.Pointer(foo.data2), defaultAlign(int(foo.size2)))
	p.Put(unsafe.Pointer(foo.data3), defaultAlign(int(foo.size3)))

	C.free(fooptr)
}

func benchmarkHandleArray(p Pool) {
	h := cgo.NewHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	fooptr := unsafe.Pointer(C.alloc_bytepool(ctx))
	foo := (*C.foo_t)(fooptr)

	data1arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data1))
	data1 := data1arr[:int(foo.size1):defaultAlign(int(foo.size1))]
	data2arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data2))
	data2 := data2arr[:int(foo.size2):defaultAlign(int(foo.size2))]
	data3arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data3))
	data3 := data3arr[:int(foo.size3):defaultAlign(int(foo.size3))]

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(unsafe.Pointer(foo.data1), defaultAlign(int(foo.size1)))
	p.Put(unsafe.Pointer(foo.data2), defaultAlign(int(foo.size2)))
	p.Put(unsafe.Pointer(foo.data3), defaultAlign(int(foo.size3)))

	C.free(fooptr)
}

func benchmarkMalloc() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	data1 := C.GoBytes(unsafe.Pointer(foo.data1), C.int(foo.size1))
	data2 := C.GoBytes(unsafe.Pointer(foo.data2), C.int(foo.size2))
	data3 := C.GoBytes(unsafe.Pointer(foo.data3), C.int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkMallocReflect() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = defaultAlign(int(foo.size1))
	s1.Len = int(foo.size1)
	s1.Data = uintptr(unsafe.Pointer(foo.data1))

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = defaultAlign(int(foo.size2))
	s2.Len = int(foo.size2)
	s2.Data = uintptr(unsafe.Pointer(foo.data2))

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = defaultAlign(int(foo.size3))
	s3.Len = int(foo.size3)
	s3.Data = uintptr(unsafe.Pointer(foo.data3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkGo(p Pool) {
	n1 := 16 * 1024
	n2 := 4 * 1024
	n3 := 512
	a1 := defaultAlign(n1)
	a2 := defaultAlign(n2)
	a3 := defaultAlign(n3)
	ptr1 := p.Get(n1)
	ptr2 := p.Get(n2)
	ptr3 := p.Get(n3)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = a1
	s1.Len = n1
	s1.Data = uintptr(ptr1)

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = a2
	s2.Len = n2
	s2.Data = uintptr(ptr2)

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = a3
	s3.Len = n3
	s3.Data = uintptr(ptr3)

	data1[0] = 110
	data1[1000] = 111
	data1[16383] = 112
	data2[0] = 120
	data2[500] = 121
	data2[4095] = 122
	data3[0] = 130
	data3[510] = 131
	data3[511] = 132

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(ptr1, n1)
	p.Put(ptr2, n2)
	p.Put(ptr3, n3)
}
