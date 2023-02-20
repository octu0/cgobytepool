# `cgobytepool`

An implementation of `[]byte` pool that can be shared between C/C++ and Go using [cgo.Handle](https://pkg.go.dev/runtime/cgo#Handle)

# How to use

Memory allocated by Go must be [C.CBytes](https://pkg.go.dev/cmd/cgo)

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

func ExampleGo(p Pool) {
	data := p.Get(100)
	defer p.Put(data)

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
cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
BenchmarkCgoBytePool
BenchmarkCgoBytePool/cgohandle
BenchmarkCgoBytePool/cgohandle-8         	   98191	       11206 ns/op	   50204 B/op	      16 allocs/op
BenchmarkCgoBytePool/cgohandle_reflect
BenchmarkCgoBytePool/cgohandle_reflect-8 	  370576	        6569 ns/op	    5108 B/op	       9 allocs/op
BenchmarkCgoBytePool/cgohandle_array
BenchmarkCgoBytePool/cgohandle_array-8   	  126110	        9421 ns/op	    5100 B/op	       9 allocs/op
BenchmarkCgoBytePool/malloc
BenchmarkCgoBytePool/malloc-8            	  395970	        3104 ns/op	   21008 B/op	       4 allocs/op
BenchmarkCgoBytePool/malloc_reflect
BenchmarkCgoBytePool/malloc_reflect-8    	10014538	       129.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/cgohandle/go
BenchmarkCgoBytePool/cgohandle/go-8      	 3484324	       310.3 ns/op	       0 B/op	       0 allocs/op
PASS
```

# License

MIT, see LICENSE file for details.
