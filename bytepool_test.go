package cgobytepool

import (
	"testing"
	"unsafe"
)

func TestCMallocPool(t *testing.T) {
	t.Run("AllocBytes", func(tt *testing.T) {
		poolSize := 10
		p := newCMallocPool(poolSize, 100)
		defer p.Close()

		ptr1 := p.Get()
		ptr2 := p.Get()
		ptr3 := p.Get()
		if p.AllocBytes() != 300 {
			tt.Errorf("3 item(100*3) get actual=%d", p.AllocBytes())
		}
		p.Put(ptr1, 100)
		p.Put(ptr2, 100)
		p.Put(ptr3, 100)
		if p.AllocBytes() != 300 {
			tt.Errorf("put pool all items actual=%d", p.AllocBytes())
		}

		ptrs := make([]unsafe.Pointer, poolSize+1)
		for i := 0; i < poolSize; i += 1 {
			ptrs[i] = p.Get()
		}
		if p.AllocBytes() != 1000 {
			tt.Errorf("reuse 300 + new 700, actual=%d", p.AllocBytes())
		}
		ptrs[10] = p.Get()
		if p.AllocBytes() != 1100 {
			tt.Errorf("new 1 item, actual=%d", p.AllocBytes())
		}
		for i := 0; i < poolSize; i += 1 {
			p.Put(ptrs[i], 100)
		}

		if p.AllocBytes() != 1100 {
			tt.Errorf("put all item but 1 item used, actual=%d", p.AllocBytes())
		}
		p.Put(ptrs[10], 100)
		if p.AllocBytes() != 1000 {
			tt.Errorf("1 item freed, actual=%d", p.AllocBytes())
		}
	})
	t.Run("Get/Put", func(tt *testing.T) {
		p := newCMallocPool(1, 100)
		defer p.Close()

		ptr1 := p.Get()
		if ptr1 == nil {
			tt.Fatalf("must alloc")
		}
		data1 := unsafe.Slice((*byte)(ptr1), 100)
		tt.Logf("----- %T", data1)
		for i := 0; i < len(data1); i += 1 {
			data1[i] = 123
		}
		p.Put(ptr1, 100)

		ptr2 := p.Get() // reuse
		data2 := unsafe.Slice((*byte)(ptr2), 100)
		for i := 0; i < len(data2); i += 1 {
			if data2[i] != 123 {
				tt.Errorf("reuse data data2[%d]=%d", i, data2[i])
			}
		}
		p.Put(ptr2, 100)
	})
}

func TestCgoBytePool(t *testing.T) {
	t.Run("MemoryAlignmentFunc", func(tt *testing.T) {
		if DefaultMemoryAlignmentFunc(500) != 752 {
			tt.Errorf("alignmentsize = %d, 500 + align = %d", defaultMemoryAlignmentSize, DefaultMemoryAlignmentFunc(500))
		}
		a8 := func(n int) int {
			return ((n + 8) >> 3) << 3
		}
		if a8(500) != 504 {
			tt.Errorf("alignmentsize = %d, 500 + align = %d", 8, a8(500))
		}
		a16 := func(n int) int {
			return ((n + 16) >> 3) << 3
		}
		if a16(500) != 512 {
			tt.Errorf("alignmentsize = %d, 500 + align = %d", 16, a16(500))
		}
	})
	t.Run("fallback", func(tt *testing.T) {
		p := NewPool(DefaultMemoryAlignmentFunc, WithPoolSize(2, 100))

		ptr1 := p.Get(500)
		if p.AllocBytes() != 752 {
			tt.Errorf("fallback alloc actual=%d", p.AllocBytes())
		}
		ptr2 := p.Get(100)
		if p.TotalAllocBytes() != 1104 {
			tt.Errorf("pools alloc actual=%d", p.TotalAllocBytes())
		}

		p.Put(ptr1, 500)
		if p.AllocBytes() != 0 {
			tt.Errorf("fallback put actual=%d", p.AllocBytes())
		}
		if p.TotalAllocBytes() != 352 {
			tt.Errorf("ptr2 is active actual=%d", p.TotalAllocBytes())
		}
		p.Put(ptr2, 100)
		p.Close()
		if p.TotalAllocBytes() != 0 {
			tt.Errorf("all item put actual=%d", p.TotalAllocBytes())
		}
	})
	t.Run("multipool", func(tt *testing.T) {
		p := NewPool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1, 100),
			WithPoolSize(1, 200),
			WithPoolSize(1, 300),
		)
		defer p.Close()

		ptr1 := p.Get(100)
		if p.pools[0].AllocBytes() != 352 {
			tt.Errorf("in pool alloc actual=%d", p.pools[0].AllocBytes())
		}
		ptr2 := p.Get(200)
		if p.pools[1].AllocBytes() != 456 {
			tt.Errorf("in pool alloc actual=%d", p.pools[1].AllocBytes())
		}
		ptr3 := p.Get(300)
		if p.pools[2].AllocBytes() != 552 {
			tt.Errorf("in pool alloc actual=%d", p.pools[2].AllocBytes())
		}
		ptr11 := p.Get(100)
		if p.pools[0].AllocBytes() != 704 {
			tt.Errorf("in pool alloc actual=%d", p.pools[0].AllocBytes())
		}
		ptr12 := p.Get(200)
		if p.pools[1].AllocBytes() != 912 {
			tt.Errorf("in pool alloc actual=%d", p.pools[1].AllocBytes())
		}
		ptr13 := p.Get(300)
		if p.pools[2].AllocBytes() != 1104 {
			tt.Errorf("in pool alloc actual=%d", p.pools[2].AllocBytes())
		}

		p.Put(ptr1, 100)
		if p.pools[0].AllocBytes() != 704 {
			tt.Errorf("in pool alloc actual=%d", p.pools[0].AllocBytes())
		}
		p.Put(ptr2, 200)
		if p.pools[1].AllocBytes() != 912 {
			tt.Errorf("in pool alloc actual=%d", p.pools[1].AllocBytes())
		}
		p.Put(ptr3, 300)
		if p.pools[2].AllocBytes() != 1104 {
			tt.Errorf("in pool alloc actual=%d", p.pools[2].AllocBytes())
		}

		p.Put(ptr11, 100)
		if p.pools[0].AllocBytes() != 352 {
			tt.Errorf("in pool alloc actual=%d", p.pools[0].AllocBytes())
		}
		p.Put(ptr12, 200)
		if p.pools[1].AllocBytes() != 456 {
			tt.Errorf("in pool alloc actual=%d", p.pools[1].AllocBytes())
		}
		p.Put(ptr13, 300)
		if p.pools[2].AllocBytes() != 552 {
			tt.Errorf("in pool alloc actual=%d", p.pools[2].AllocBytes())
		}
	})
}
