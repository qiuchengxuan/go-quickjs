package quickjs

//#include "ffi.h"
import "C"
import (
	"unsafe"
)

type Signed interface{ int8 | int16 | int32 }
type Unsigned interface{ uint8 | uint16 | uint32 }
type Float interface{ float32 | float64 }
type Number interface{ Signed | Unsigned | Float }

const (
	typedArrayUInt8C = iota
	typedArrayInt8
	typedArrayUint8
	typedArrayInt16
	typedArrayUint16
	typedArrayInt32
	typedArrayUint32
	typedArrayBigInt64
	typedArrayBigUint64
	typedArrayFloat32
	typedArrayFloat64
)

func newTypedArray[T Number](c *Context, slice []T, arrayType int) C.JSValue {
	arrayBuf := c.assert(C.JS_NewArrayBufferCopy(c.raw, slicePtr(slice), sliceSize(slice)))
	retval := C.JS_NewTypedArray(c.raw, C.int(1), &arrayBuf, C.JSTypedArrayEnum(arrayType))
	return c.assert(retval)
}

type Array struct{ Object }

func (a Array) Len() int {
	property, _ := a.GetProperty("length")
	return property.ToPrimitive().(int)
}

func (a Array) Set(index int, value any) {
	a.SetPropertyByIndex(uint32(index), value)
}

func (a Array) Get(index int) Value {
	return a.GetPropertyByIndex(uint32(index))
}

func (a Array) ToNative() []any {
	length := a.Len()
	retval := make([]any, 0, length)
	for i := 0; i < length; i++ {
		retval = append(retval, a.Get(i).ToNative())
	}
	return retval
}

type ArrayBuffer struct{ Object }

func (b ArrayBuffer) Len() int {
	property, err := b.GetProperty("byteLength")
	if err != nil {
		panic(err)
	}
	return property.ToNative().(int)
}

func (b ArrayBuffer) ToNative() []byte {
	size := C.size_t(b.Len())
	out := C.JS_GetArrayBuffer(b.context.raw, &size, b.raw)
	return C.GoBytes(unsafe.Pointer(out), C.int(size))
}

type TypedArray[T Number] struct{ Object }

func (a TypedArray[T]) Len() int {
	property, _ := a.GetProperty("length")
	return property.ToNative().(int)
}

func (a TypedArray[T]) ToNative() []T {
	buf := C.JS_GetTypedArrayBuffer(a.context.raw, a.raw, nil, nil, nil)
	bytes := ArrayBuffer{Value{a.context, buf}.Object()}.ToNative()
	C.JS_FreeValue(a.context.raw, buf)
	var t T
	sizeOf := int(unsafe.Sizeof(t))
	return unsafe.Slice((*T)(unsafe.Pointer(&bytes[0])), len(bytes)/sizeOf)
}

// Assume object is Array
func (o Object) Array() Array { return Array{o} }

// Assume object is ArrayBuffer
func (o Object) ArrayBuffer() ArrayBuffer { return ArrayBuffer{o} }
