package cgobytepool

import (
	"testing"

	"github.com/octu0/bp"
)

func BenchmarkCgoBytePool(b *testing.B) {
	b.Run("cgohandle", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkHandle(p)
			}
		})
	})
	b.Run("cgohandle_reflect", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkHandleReflect(p)
			}
		})
	})
	b.Run("cgohandle_array", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkHandleArray(p)
			}
		})
	})
	b.Run("cgohandle_unsafeslice", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkHandleUnsafeSlice(p)
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
	b.Run("malloc_unsafeslice", func(tb *testing.B) {
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkMallocUnsafeSlice()
			}
		})
	})
	b.Run("go/malloc", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				benchmarkGo(p)
			}
		})
	})
	b.Run("go/cgo", func(tb *testing.B) {
		p := NewCgoBytePool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1000, 16*1024),
			WithPoolSize(1000, 4*1024),
			WithPoolSize(1000, 512),
		)
		tb.ResetTimer()
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
	b.Run("bp", func(tb *testing.B) {
		poolSize := 1000
		p := bp.NewMultiBytePool(
			bp.MultiBytePoolSize(poolSize, 16*1024),
			bp.MultiBytePoolSize(poolSize, 4*1024),
			bp.MultiBytePoolSize(poolSize, 512),
		)
		tb.ResetTimer()
		tb.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				data1 := p.Get(16 * 1024)
				data2 := p.Get(4 * 1024)
				data3 := p.Get(512)

				p.Put(data1)
				p.Put(data2)
				p.Put(data3)
			}
		})
	})
}
