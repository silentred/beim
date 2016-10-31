package lib

import (
	"reflect"
	"unsafe"
)

// ByteString converts byte to string
func ByteString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringByte converts string to byte.
// DANDER: SHOULD NOT read or write cap of the slice
func StringByte(s string) []byte {
	//return []byte(s)

	//return *(*[]byte)(unsafe.Pointer(&s))

	var sh reflect.SliceHeader
	sh.Len = len(s)
	sh.Cap = len(s)
	sh.Data = uintptr(stringPointer(s))
	return *(*[]byte)(unsafe.Pointer(&sh))
}

func stringPointer(s string) unsafe.Pointer {
	sHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return unsafe.Pointer(sHeader.Data)
}

func bytesPointer(b []byte) unsafe.Pointer {
	sHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return unsafe.Pointer(sHeader.Data)
}
