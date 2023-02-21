# `cgobytepool`

Shared byte pool implementation between C and Go(cgo)  
`unsigned char*` in C / `[]byte` in Go (convertible using [unsafe.Slice](https://pkg.go.dev/unsafe#Slice) or [reflect.SliceHeader](https://pkg.go.dev/reflect#SliceHeader))  
this pool shares using [cgo.Handle](https://pkg.go.dev/runtime/cgo#Handle)

# How to use

Need to declare `extern` in C and declare `export` in Go

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
	println(len(data)) // => 100
	println(cap(data)) // => 100

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
```

# Benchmark

```
goos: darwin
goarch: amd64
pkg: github.com/octu0/cgobytepool/benchmark
cpu: Intel(R) Core(TM) i7-8569U CPU @ 2.80GHz
BenchmarkCgoBytePool
BenchmarkCgoBytePool/cgohandle
BenchmarkCgoBytePool/cgohandle-8         	              219858	       4697 ns/op	     21205 B/op	      10 allocs/op
BenchmarkCgoBytePool/cgohandle_reflect
BenchmarkCgoBytePool/cgohandle_reflect-8 	              571408	       2064 ns/op	       198 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgohandle_array
BenchmarkCgoBytePool/cgohandle_array-8   	              583810	       2048 ns/op	       197 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgohandle_unsafeslice
BenchmarkCgoBytePool/cgohandle_unsafeslice-8         	  597548	       2298 ns/op	       199 B/op	       6 allocs/op
BenchmarkCgoBytePool/malloc
BenchmarkCgoBytePool/malloc-8                        	  347733	       2967 ns/op	     21008 B/op	       4 allocs/op
BenchmarkCgoBytePool/malloc_reflect
BenchmarkCgoBytePool/malloc_reflect-8                	10762393	        120.6 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/malloc_unsafeslice
BenchmarkCgoBytePool/malloc_unsafeslice-8            	 8975125	        138.8 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/go/malloc
BenchmarkCgoBytePool/go/malloc-8                     	 3332970	        325.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkCgoBytePool/go/cgo
BenchmarkCgoBytePool/go/cgo-8                        	  967604	       1178 ns/op	        95 B/op	       3 allocs/op
BenchmarkCgoBytePool/bp
BenchmarkCgoBytePool/bp-8                            	 4372924	        305.3 ns/op	       0 B/op	       0 allocs/op
PASS
```

# License

MIT, see LICENSE file for details.
