package rgo

import (
	"reflect"
	"unsafe"
)

func eqSimpleData(a, b SimpleData) bool {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	aSize := int(av.Type().Size())
	bSize := int(bv.Type().Size())
	if aSize != bSize {
		return false
	}
	aBytes := (*[1 << 30]byte)(unsafe.Pointer(av.UnsafeAddr()))[:aSize]
	bBytes := (*[1 << 30]byte)(unsafe.Pointer(bv.UnsafeAddr()))[:bSize]
	for i := range aBytes {
		if aBytes[i] != bBytes[i] {
			return false
		}
	}
	return true
}
