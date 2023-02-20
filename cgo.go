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

static void write_foo(foo_t *foo) {
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
  write_foo(foo);
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
  write_foo(foo);
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
	"runtime/cgo"
	"unsafe"
)

type Pool interface {
	Get(int) []byte
	Put([]byte)
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
	data := p.Get(defaultAlign(n))
	return C.CBytes(data[:n]) // cgo.calloc?
}

//export bytepool_put
func bytepool_put(ctx unsafe.Pointer, data unsafe.Pointer, size C.size_t) {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := defaultAlign(int(size))
	b := C.GoBytes(data, C.int(n))
	p.Put(b[:n])
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
	pools []*byteslicepool // github.com/octu0/bp.(*BytePool)
}

func (p *defaultPool) find(size int) (*byteslicepool, bool) {
	for _, pp := range p.pools {
		if size <= pp.bufSize {
			return pp, true
		}
	}
	return nil, false
}

func (p *defaultPool) Get(n int) []byte {
	if pp, ok := p.find(n); ok {
		return pp.Get()[:n]
	}
	return make([]byte, n) // fallback
}

func (p *defaultPool) Put(b []byte) {
	n := cap(b)
	if pp, ok := p.find(n); ok {
		pp.Put(b)
	}
}

func NewDefaultPool() *defaultPool {
	poolSize := 1000
	pools := make([]*byteslicepool, 7)
	for i := 0; i < 7; i += 1 {
		pools[i] = newByteSlicePool(poolSize, defaultAlign(4096*(i+1)))
	}
	return &defaultPool{pools}
}

type byteslicepool struct {
	pool    chan []byte
	bufSize int
}

func (p *byteslicepool) Get() []byte {
	select {
	case buf := <-p.pool:
		// reuse
		return buf[:p.bufSize]
	default:
		// new
		return make([]byte, p.bufSize)
	}
}

func (p *byteslicepool) Put(data []byte) {
	if cap(data) < p.bufSize {
		return // discard
	}
	select {
	case p.pool <- data[:p.bufSize]:
		// ok
	default:
		// release
	}
}

func newByteSlicePool(poolSize, bufSize int) *byteslicepool {
	return &byteslicepool{
		pool:    make(chan []byte, poolSize),
		bufSize: bufSize,
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

	p.Put(data1)
	p.Put(data2)
	p.Put(data3)

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

	p.Put(data1)
	p.Put(data2)
	p.Put(data3)

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
	data1 := p.Get(16 * 1024)
	data2 := p.Get(4 * 1024)
	data3 := p.Get(512)
	p.Put(data1)
	p.Put(data2)
	p.Put(data3)
}
