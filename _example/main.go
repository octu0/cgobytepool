package main

/*
#include <stdlib.h>

extern void *bytepool_get(void *context, size_t size);
extern void bytepool_put(void *context, void *data, size_t size);
extern void bytepool_free(void *context);

static void do_something(unsigned char *data) {
  // nop
}

static void ExampleCgo(void *ctx) {
  unsigned char *data = (unsigned char*) bytepool_get(ctx, 100);
  do_something(data);
  bytepool_put(ctx, data, 100);
  bytepool_free(ctx);
}
*/
import "C"

import (
	"unsafe"

	"github.com/octu0/cgobytepool"
)

//export bytepool_get
func bytepool_get(ctx unsafe.Pointer, size C.size_t) unsafe.Pointer {
	return cgobytepool.HandlePoolGet(ctx, int(size))
}

//export bytepool_put
func bytepool_put(ctx unsafe.Pointer, data unsafe.Pointer, size C.size_t) {
	cgobytepool.HandlePoolPut(ctx, data, int(size))
}

//export bytepool_free
func bytepool_free(ctx unsafe.Pointer) {
	cgobytepool.HandlePoolFree(ctx)
}

func ExampleGo(p cgobytepool.Pool) {
	ptr := p.Get(100)
	defer p.Put(ptr, 100)

	data := unsafe.Slice((*byte)(ptr), 100)
	println(len(data))
	println(cap(data))

	doSomething(data)
}

func main() {
	p := cgobytepool.NewPool(
		cgobytepool.DefaultMemoryAlignmentFunc,
		cgobytepool.WithPoolSize(1000, 16*1024),
		cgobytepool.WithPoolSize(1000, 4*1024),
		cgobytepool.WithPoolSize(1000, 512),
	)
	ExampleGo(p)

	h := cgobytepool.CgoHandle(p)

	C.ExampleCgo(unsafe.Pointer(&h))
}

func doSomething(p []byte) {
	// nop
}
