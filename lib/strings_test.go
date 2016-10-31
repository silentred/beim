package lib

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func Benchmark_ByteString(b *testing.B) {
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		_ = string(bs)
	}
}

func Benchmark_ByteStringPointer(b *testing.B) {
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		_ = ByteString(bs)
	}
}

var out []byte

func Benchmark_StringByte(b *testing.B) {
	s := "test"
	for i := 0; i < b.N; i++ {
		out = StringByte(s)
	}
}

func Benchmark_StringByte_Simple(b *testing.B) {
	s := "test"
	for i := 0; i < b.N; i++ {
		out = []byte(s)
	}
}

func Benchmark_StringByte_Quick(b *testing.B) {
	s := "test"
	for i := 0; i < b.N; i++ {
		out = quickStringByte(s)
	}
}

func quickStringByte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func TestPointer(t *testing.T) {
	s := "testgo"
	sHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	fmt.Printf("ptr is %p %p, %p \n", &s, sHeader, &sHeader.Data)
	l := (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(sHeader)) + 8))
	fmt.Println("len is", sHeader.Len, *l)

	var sh reflect.SliceHeader
	fmt.Printf("%p \n", &sh)

}
