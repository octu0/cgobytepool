package cgobytepool

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestGoBytes(t *testing.T) {
	t.Run("GoBytes", func(tt *testing.T) {
		data := []byte{100, 200, 123}
		out := GoBytes(unsafe.Pointer(&data[0]), 3)
		if (out[0] == 100 && out[1] == 200 && out[2] == 123) != true {
			tt.Errorf("actual=%v", out)
		}
	})
	t.Run("ReflectGoBytes", func(tt *testing.T) {
		data := []int16{32767, -32768}
		out := ReflectGoBytes(unsafe.Pointer(&data[0]), 4, 4)

		v1 := int16(0)
		if err := binary.Read(bytes.NewReader(out[0:2]), binary.LittleEndian, &v1); err != nil {
			tt.Fatalf("no error: %+v", err)
		}
		v2 := int16(0)
		if err := binary.Read(bytes.NewReader(out[2:4]), binary.LittleEndian, &v2); err != nil {
			tt.Fatalf("no error: %+v", err)
		}
		if v1 != 32767 {
			tt.Errorf("actual=%v", v1)
		}
		if v2 != -32768 {
			tt.Errorf("actual=%v", v2)
		}
	})
}

func TestCBytes(t *testing.T) {
	t.Run("CBytes", func(tt *testing.T) {
		p := NewPool(
			DefaultMemoryAlignmentFunc,
			WithPoolSize(1, 100),
			WithPoolSize(1, 200),
			WithPoolSize(1, 300),
		)
		defer p.Close()

		data := []byte("hello world")
		ptr, done := CBytes(p, data)
		defer done()

		if p.TotalAllocBytes() != 352 { // DefaultMemoryAlignmentSize = 256; ((100 + 256) >> 3) << 3 is 352
			tt.Errorf("alloc data actual=%d", p.TotalAllocBytes())
		}

		out := GoBytes(ptr, len(data))
		if bytes.Equal(out, data) != true {
			tt.Errorf("actual=%v", out)
		}
	})
}
