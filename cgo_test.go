package cgobytepool

import (
	"testing"
)

func BenchmarkCgoBytePool(b *testing.B) {
	b.Run("cgohandle", func(tb *testing.B) {
		benchmarkHandle(tb.N)
	})
	b.Run("malloc", func(tb *testing.B) {
		benchmarkHandle(tb.N)
	})
	b.Run("cgohandle/go", func(tb *testing.B) {
		benchmarkHandleAndGo(tb.N)
	})
}
