package benchmark

/*
#include <stdlib.h>
#include <string.h>

extern void *bytepool_get(void *context, size_t size);
extern void bytepool_put(void *context, void *data, size_t size);
extern void bytepool_free(void *context);

typedef struct foo_t {
  unsigned char *data1;
  unsigned char *data2;
  unsigned char *data3;
  int size1;
  int size2;
  int size3;
} foo_t;

static void test_write_foo(foo_t *foo) {
  foo->data1[0] = 110;
  foo->data1[1000] = 111;
  foo->data1[16383] = 112;
  foo->data2[0] = 120;
  foo->data2[500] = 121;
  foo->data2[4095] = 122;
  foo->data3[0] = 130;
  foo->data3[510] = 131;
  foo->data3[511] = 132;
}

static foo_t *alloc_bytepool(void *ctx) {
  foo_t *foo = (foo_t *) malloc(sizeof(foo_t));
  memset(foo, 0, sizeof(foo_t));

  int size1 = 16 * 1024;
  int size2 = 4 * 1024;
  int size3 = 512;
  foo->data1 = (unsigned char *) bytepool_get(ctx, size1);
  foo->size1 = size1;
  foo->data2 = (unsigned char *) bytepool_get(ctx, size2);
  foo->size2 = size2;
  foo->data3 = (unsigned char *) bytepool_get(ctx, size3);
  foo->size3 = size3;
  test_write_foo(foo);
  return foo;
}

static void free_bytepool(void *ctx, foo_t *foo) {
  if(foo != NULL) {
    bytepool_put(ctx, foo->data1, foo->size1);
    bytepool_put(ctx, foo->data2, foo->size2);
    bytepool_put(ctx, foo->data3, foo->size3);
  }
  free(foo);
}

static foo_t *alloc_malloc() {
  foo_t *foo = (foo_t *) malloc(sizeof(foo_t));
  memset(foo, 0, sizeof(foo_t));

  int size1 = 16 * 1024;
  int size2 = 4 * 1024;
  int size3 = 512;
  foo->data1 = (unsigned char *) malloc(size1);
  foo->size1 = size1;
  foo->data2 = (unsigned char *) malloc(size2);
  foo->size2 = size2;
  foo->data3 = (unsigned char *) malloc(size3);
  foo->size3 = size3;
  test_write_foo(foo);
  return foo;
}

static void free_malloc(foo_t *foo) {
  if(foo != NULL) {
    free(foo->data1);
    free(foo->data2);
    free(foo->data3);
  }
  free(foo);
}

static int check1(unsigned char *data) {
  if (data[0] == 140 && data[1000] == 141 && data[16383] == 142) {
    return 0;
  }
  return -1;
}

static int check2(unsigned char *data) {
  if (data[0] == 150 && data[500] == 151 && data[4095] == 152) {
    return 0;
  }
  return -1;
}

static int check3(unsigned char *data) {
  if (data[0] == 160 && data[510] == 161 && data[511] == 162) {
    return 0;
  }
  return -1;
}
*/
import "C"

import (
	"reflect"
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

func checkCdata1(data []byte) {
	if len(data) != 16384 {
		panic("data1 length 16384")
	}
	if data[0] != 110 {
		panic("data1[0] 110")
	}
	if data[1000] != 111 {
		panic("data1[1000] 111")
	}
	if data[16383] != 112 {
		panic("data1[16383] 112")
	}
}

func checkCdata2(data []byte) {
	if len(data) != 4096 {
		panic("data2 length 4096")
	}
	if data[0] != 120 {
		panic("data2[0] 120")
	}
	if data[500] != 121 {
		panic("data2[500] 121")
	}
	if data[4095] != 122 {
		panic("data2[4095] 122")
	}
}

func checkCdata3(data []byte) {
	if len(data) != 512 {
		panic("data3 length 512")
	}
	if data[0] != 130 {
		panic("data3[0] 130")
	}
	if data[510] != 131 {
		panic("data3[510] 131")
	}
	if data[511] != 132 {
		panic("data3[511] 132")
	}
}

func benchmarkHandle(p cgobytepool.Pool) {
	h := cgobytepool.CgoHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_bytepool(ctx)))
	defer C.free_bytepool(ctx, foo)

	data1 := C.GoBytes(unsafe.Pointer(foo.data1), C.int(foo.size1))
	data2 := C.GoBytes(unsafe.Pointer(foo.data2), C.int(foo.size2))
	data3 := C.GoBytes(unsafe.Pointer(foo.data3), C.int(foo.size3))
	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkHandleReflect(p cgobytepool.Pool) {
	h := cgobytepool.CgoHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	fooptr := unsafe.Pointer(C.alloc_bytepool(ctx))
	foo := (*C.foo_t)(fooptr)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = int(foo.size1)
	s1.Len = int(foo.size1)
	s1.Data = uintptr(unsafe.Pointer(foo.data1))

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = int(foo.size2)
	s2.Len = int(foo.size2)
	s2.Data = uintptr(unsafe.Pointer(foo.data2))

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = int(foo.size3)
	s3.Len = int(foo.size3)
	s3.Data = uintptr(unsafe.Pointer(foo.data3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(unsafe.Pointer(foo.data1), int(foo.size1))
	p.Put(unsafe.Pointer(foo.data2), int(foo.size2))
	p.Put(unsafe.Pointer(foo.data3), int(foo.size3))

	C.free(fooptr)
}

func benchmarkHandleArray(p cgobytepool.Pool) {
	h := cgobytepool.CgoHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	fooptr := unsafe.Pointer(C.alloc_bytepool(ctx))
	foo := (*C.foo_t)(fooptr)

	data1arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data1))
	data1 := data1arr[:int(foo.size1):int(foo.size1)]
	data2arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data2))
	data2 := data2arr[:int(foo.size2):int(foo.size2)]
	data3arr := (*[1 << 32]byte)(unsafe.Pointer(foo.data3))
	data3 := data3arr[:int(foo.size3):int(foo.size3)]

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(unsafe.Pointer(foo.data1), int(foo.size1))
	p.Put(unsafe.Pointer(foo.data2), int(foo.size2))
	p.Put(unsafe.Pointer(foo.data3), int(foo.size3))

	C.free(fooptr)
}

func benchmarkHandleUnsafeSlice(p cgobytepool.Pool) {
	h := cgobytepool.CgoHandle(p)
	defer h.Delete()

	ctx := unsafe.Pointer(&h)
	fooptr := unsafe.Pointer(C.alloc_bytepool(ctx))
	foo := (*C.foo_t)(fooptr)

	data1 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data1)), int(foo.size1))
	data2 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data2)), int(foo.size2))
	data3 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data3)), int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(unsafe.Pointer(foo.data1), int(foo.size1))
	p.Put(unsafe.Pointer(foo.data2), int(foo.size2))
	p.Put(unsafe.Pointer(foo.data3), int(foo.size3))

	C.free(fooptr)
}

func benchmarkMalloc() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	data1 := C.GoBytes(unsafe.Pointer(foo.data1), C.int(foo.size1))
	data2 := C.GoBytes(unsafe.Pointer(foo.data2), C.int(foo.size2))
	data3 := C.GoBytes(unsafe.Pointer(foo.data3), C.int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkMallocReflect() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = int(foo.size1)
	s1.Len = int(foo.size1)
	s1.Data = uintptr(unsafe.Pointer(foo.data1))

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = int(foo.size2)
	s2.Len = int(foo.size2)
	s2.Data = uintptr(unsafe.Pointer(foo.data2))

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = int(foo.size3)
	s3.Len = int(foo.size3)
	s3.Data = uintptr(unsafe.Pointer(foo.data3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkMallocReflect2() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	data1 := cgobytepool.GoBytes(unsafe.Pointer(foo.data1), int(foo.size1))
	data2 := cgobytepool.GoBytes(unsafe.Pointer(foo.data2), int(foo.size2))
	data3 := cgobytepool.GoBytes(unsafe.Pointer(foo.data3), int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkMallocUnsafeSlice() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	data1 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data1)), int(foo.size1))
	data2 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data2)), int(foo.size2))
	data3 := unsafe.Slice((*byte)(unsafe.Pointer(foo.data3)), int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkMallocUnsafeSlice2() {
	foo := (*C.foo_t)(unsafe.Pointer(C.alloc_malloc()))
	defer C.free_malloc(foo)

	data1 := cgobytepool.UnsafeGoBytes(unsafe.Pointer(foo.data1), int(foo.size1))
	data2 := cgobytepool.UnsafeGoBytes(unsafe.Pointer(foo.data2), int(foo.size2))
	data3 := cgobytepool.UnsafeGoBytes(unsafe.Pointer(foo.data3), int(foo.size3))

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)
}

func benchmarkGo(p cgobytepool.Pool) {
	n1 := 16 * 1024
	n2 := 4 * 1024
	n3 := 512
	ptr1 := p.Get(n1)
	ptr2 := p.Get(n2)
	ptr3 := p.Get(n3)

	var data1, data2, data3 []byte
	s1 := (*reflect.SliceHeader)(unsafe.Pointer(&data1))
	s1.Cap = n1
	s1.Len = n1
	s1.Data = uintptr(ptr1)

	s2 := (*reflect.SliceHeader)(unsafe.Pointer(&data2))
	s2.Cap = n2
	s2.Len = n2
	s2.Data = uintptr(ptr2)

	s3 := (*reflect.SliceHeader)(unsafe.Pointer(&data3))
	s3.Cap = n3
	s3.Len = n3
	s3.Data = uintptr(ptr3)

	data1[0] = 110
	data1[1000] = 111
	data1[16383] = 112
	data2[0] = 120
	data2[500] = 121
	data2[4095] = 122
	data3[0] = 130
	data3[510] = 131
	data3[511] = 132

	checkCdata1(data1)
	checkCdata2(data2)
	checkCdata3(data3)

	p.Put(ptr1, n1)
	p.Put(ptr2, n2)
	p.Put(ptr3, n3)
}

func benchmarkCBytes() {
	data1 := make([]byte, 16*1024)
	data1[0] = 140
	data1[1000] = 141
	data1[16383] = 142

	data2 := make([]byte, 4*1024)
	data2[0] = 150
	data2[500] = 151
	data2[4095] = 152

	data3 := make([]byte, 512)
	data3[0] = 160
	data3[510] = 161
	data3[511] = 162

	ptr1 := C.CBytes(data1)
	ptr2 := C.CBytes(data2)
	ptr3 := C.CBytes(data3)

	if ret := C.check1((*C.uchar)(ptr1)); ret != C.int(0) {
		panic("err")
	}
	if ret := C.check2((*C.uchar)(ptr2)); ret != C.int(0) {
		panic("err")
	}
	if ret := C.check3((*C.uchar)(ptr3)); ret != C.int(0) {
		panic("err")
	}
}

func benchmarkCBytes_cgobytepool(p cgobytepool.Pool) {
	data1 := make([]byte, 16*1024)
	data1[0] = 140
	data1[1000] = 141
	data1[16383] = 142

	data2 := make([]byte, 4*1024)
	data2[0] = 150
	data2[500] = 151
	data2[4095] = 152

	data3 := make([]byte, 512)
	data3[0] = 160
	data3[510] = 161
	data3[511] = 162

	ptr1, done1 := cgobytepool.CBytes(p, data1)
	defer done1()
	ptr2, done2 := cgobytepool.CBytes(p, data2)
	defer done2()
	ptr3, done3 := cgobytepool.CBytes(p, data3)
	defer done3()

	if ret := C.check1((*C.uchar)(ptr1)); ret != C.int(0) {
		panic("err")
	}
	if ret := C.check2((*C.uchar)(ptr2)); ret != C.int(0) {
		panic("err")
	}
	if ret := C.check3((*C.uchar)(ptr3)); ret != C.int(0) {
		panic("err")
	}
}
