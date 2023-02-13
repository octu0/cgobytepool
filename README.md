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
cpu: Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz
BenchmarkCgoBytePool
BenchmarkCgoBytePool/cgohandle
BenchmarkCgoBytePool/cgohandle-4         	 1606730	       751.6 ns/op	     160 B/op	       3 allocs/op
BenchmarkCgoBytePool/malloc
BenchmarkCgoBytePool/malloc-4            	 1594532	       789.9 ns/op	     160 B/op	       3 allocs/op
BenchmarkCgoBytePool/cgohandle/go
BenchmarkCgoBytePool/cgohandle/go-4      	 2638782	       452.1 ns/op	      92 B/op	       2 allocs/op
PASS
ok  	github.com/octu0/cgobytepool	5.998s
```

# License

MIT, see LICENSE file for details.
