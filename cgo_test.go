package cgobytepool

import (
	"testing"
)

func BenchmarkCgoBytePool(b *testing.B) {
	b.Run("cgohandle", func(tb *testing.B) {
    p := NewDefaultPool()
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkHandle(p)
      }
    })
	})
	b.Run("cgohandle_reflect", func(tb *testing.B) {
    p := NewDefaultPool()
    datas := make([][]byte, 100)
    for i := 0; i < 100; i += 1 {
      datas[i] = p.Get(16 * 1024)
    }
    for i := 0; i < 100; i += 1 {
      p.Put(datas[i])
    }
    for i := 0; i < 100; i += 1 {
      datas[i] = p.Get(4 * 1024)
    }
    for i := 0; i < 100; i += 1 {
      p.Put(datas[i])
    }
    for i := 0; i < 100; i += 1 {
      datas[i] = p.Get(512)
    }
    for i := 0; i < 100; i += 1 {
      p.Put(datas[i])
    }
    tb.ResetTimer()
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkHandleReflect(p)
      }
    })
	})
	b.Run("cgohandle_array", func(tb *testing.B) {
    p := NewDefaultPool()
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkHandleArray(p)
      }
    })
	})
	b.Run("malloc", func(tb *testing.B) {
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkMalloc()
      }
    })
	})
	b.Run("malloc_reflect", func(tb *testing.B) {
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkMallocReflect()
      }
    })
	})
	b.Run("cgohandle/go", func(tb *testing.B) {
    p := NewDefaultPool()
    tb.RunParallel(func(pb *testing.PB) {
      for pb.Next() {
        benchmarkGo(p)
      }
    })
	})
}
