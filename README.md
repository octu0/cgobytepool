# `cgobytepool`

[![MIT License](https://img.shields.io/github/license/octu0/cgobytepool)](https://github.com/octu0/cgobytepool/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/octu0/cgobytepool?status.svg)](https://godoc.org/github.com/octu0/cgobytepool)
[![Go Report Card](https://goreportcard.com/badge/github.com/octu0/cgobytepool)](https://goreportcard.com/report/github.com/octu0/cgobytepool)
[![Releases](https://img.shields.io/github/v/release/octu0/cgobytepool)](https://github.com/octu0/cgobytepool/releases)

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
BenchmarkCgoBytePool/cgohandle-8         	              218674	        4764 ns/op	   21209 B/op	      10 allocs/op
BenchmarkCgoBytePool/cgohandle_reflect
BenchmarkCgoBytePool/cgohandle_reflect-8 	              572455	        2039 ns/op	     198 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgohandle_array
BenchmarkCgoBytePool/cgohandle_array-8   	              594552	        2082 ns/op	     198 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgohandle_unsafeslice
BenchmarkCgoBytePool/cgohandle_unsafeslice-8         	  580590	        2040 ns/op	     198 B/op	       6 allocs/op
BenchmarkCgoBytePool/malloc
BenchmarkCgoBytePool/malloc-8                        	  433221	        3001 ns/op	   21008 B/op	       4 allocs/op
BenchmarkCgoBytePool/malloc_reflect
BenchmarkCgoBytePool/malloc_reflect-8                	10844132	       115.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/malloc_reflect2
BenchmarkCgoBytePool/malloc_reflect2-8               	 9369612	       129.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/malloc_unsafeslice
BenchmarkCgoBytePool/malloc_unsafeslice-8            	 8921526	       133.3 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/malloc_unsafeslice2
BenchmarkCgoBytePool/malloc_unsafeslice2-8           	 8935269	       127.8 ns/op	      16 B/op	       1 allocs/op
BenchmarkCgoBytePool/go/malloc
BenchmarkCgoBytePool/go/malloc-8                     	 3910248	       334.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkCgoBytePool/go/cgo
BenchmarkCgoBytePool/go/cgo-8                        	  943117	        1166 ns/op	      96 B/op	       3 allocs/op
BenchmarkCgoBytePool/bp
BenchmarkCgoBytePool/bp-8                            	 3708741	       270.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkCgoBytePool/cbytes
BenchmarkCgoBytePool/cbytes-8                        	  231247	        5876 ns/op	   21064 B/op	       6 allocs/op
BenchmarkCgoBytePool/cgobytepool_cbytes
BenchmarkCgoBytePool/cgobytepool_cbytes-8            	 1795348	       659.6 ns/op	     144 B/op	       3 allocs/op
PASS
```

# License

MIT, see LICENSE file for details.
