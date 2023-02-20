# `cgobytepool`

An implementation of `[]byte` pool that can be shared between C/C++ and Go using [cgo.Handle](https://pkg.go.dev/runtime/cgo#Handle)

# How to use

Memory allocated by Go must be [C.malloc](https://pkg.go.dev/cmd/cgo)

```go
/*
#include <stdlib.h>

extern void *bytepool_get(void *context, size_t size);
extern void bytepool_put(void *context, void *data, size_t size);
extern void bytepool_free(void *context);

static void ExampleCgo(void *ctx) {
  unsigned char *data = (unsigned char*) bytepool_get(ctx, 100);
  do_something(data);
  bytepool_put(ctx, data, 100);
  bytepool_free(ctx);
}
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

type Pool interface {
	Get(int) unsafe.Pointer
	Put(unsafe.Pointer, int)
}

//export bytepool_get
func bytepool_get(ctx unsafe.Pointer, size C.size_t) unsafe.Pointer {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := int(size)
	return p.Get(n)
}

//export bytepool_put
func bytepool_put(ctx unsafe.Pointer, data unsafe.Pointer, size C.size_t) {
	h := *(*cgo.Handle)(ctx)

	p := h.Value().(Pool)
	n := int(size)
	p.Put(data, n)
}

//export bytepool_free
func bytepool_free(ctx unsafe.Pointer) {
	h := *(*cgo.Handle)(ctx)
	h.Delete()
}

func ExampleGo(p Pool) {
	ptr := p.Get(100)
	defer p.Put(ptr, 100)

	var data []byte
	s := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	s.Cap = 100
	s.Len = 100
	s.Data = uintptr(ptr)

	doSomething(data)
}

func main() {
	p := NewDefaultPool()
	ExampleGo(p)

	h := cgo.NewHandle(p)
	C.ExampleCgo(unsage.Pointer(&h))
}
```

# Benchmark

```
goos: darwin
goarch: amd64
pkg: github.com/octu0/cgobytepool
cpu: Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz
BenchmarkCgoBytePool
BenchmarkCgoBytePool/cgohandle
BenchmarkCgoBytePool/cgohandle-4         	  172785	        6894 ns/op	   21223 B/op	      11 allocs/op
BenchmarkCgoBytePool/cgohandle_reflect
BenchmarkCgoBytePool/cgohandle_reflect-4 	  467392	        2438 ns/op	     193 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgohandle_array
BenchmarkCgoBytePool/cgohandle_array-4   	  473598	        2492 ns/op	     196 B/op	       6 allocs/op
BenchmarkCgoBytePool/malloc
BenchmarkCgoBytePool/malloc-4            	  201094	        6285 ns/op	   21008 B/op	       4 allocs/op
BenchmarkCgoBytePool/malloc_reflect
BenchmarkCgoBytePool/malloc_reflect-4    	 3540825	        306.6 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/go/malloc
BenchmarkCgoBytePool/go/malloc-4         	 2576083	        511.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkCgoBytePool/go/cgo
BenchmarkCgoBytePool/go/cgo-4            	  732290	        1580 ns/op	      99 B/op	       3 allocs/op
PASS
```

# License

MIT, see LICENSE file for details.
