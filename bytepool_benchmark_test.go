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
	b.Run("go/malloc", func(tb *testing.B) {
		p := NewDefaultPool()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkGo(p)
			}
		})
	})
	b.Run("go/cgo", func(tb *testing.B) {
		p := NewDefaultPool()
		i := 0
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				i += 1
				if i%2 == 0 {
					benchmarkGo(p)
				} else {
					benchmarkHandleReflect(p)
				}
			}
		})
	})
}
