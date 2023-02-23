package cgobytepool

import (
	"reflect"
	"unsafe"
)

type DoneFunc func()

func GoBytes(data unsafe.Pointer, size int) []byte {
	return ReflectGoBytes(data, size, size)
}

func ReflectGoBytes(data unsafe.Pointer, sizeLen, sizeCap int) []byte {
	var out []byte

	s := (*reflect.SliceHeader)(unsafe.Pointer(&out))
	s.Cap = sizeCap
	s.Len = sizeLen
	s.Data = uintptr(data)

	return out
}

func UnsafeGoBytes(data unsafe.Pointer, size int) []byte {
	return unsafe.Slice((*byte)(data), size)
}

func CBytes(p Pool, data []byte) (unsafe.Pointer, DoneFunc) {
	size := cap(data)
	ptr := p.Get(size)
	out := ReflectGoBytes(ptr, len(data), size)
	copy(out, data)
	return ptr, func() { p.Put(ptr, size) }
}
