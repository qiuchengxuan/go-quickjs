package quickjs

//#include "ffi.h"
import "C"
import "unsafe"

type ObjectKind uint8

const (
	KindPlainObject ObjectKind = iota
	KindArray
	KindArrayBuffer
	KindInt8Array
	KindInt16Array
	KindInt32Array
	KindUint8Array
	KindUint16Array
	KindUint32Array
	KindFloat32Array
	KindFloat64Array
	KindMap
	KindSet
	KindDate
	KindUnknown
)

type Object struct{ Value }

func (o Object) Kind() ObjectKind {
	if C.JS_IsArray(o.context.raw, o.raw) == 1 {
		return KindArray
	}
	property, _ := o.GetProperty("constructor")
	name, _ := property.Object().GetProperty("name")
	switch name.String() {
	case "Object":
		return KindPlainObject
	case "ArrayBuffer":
		return KindArrayBuffer
	case "Int8Array":
		return KindInt8Array
	case "Int16Array":
		return KindInt16Array
	case "Int32Array":
		return KindInt32Array
	case "Uint8Array":
		return KindUint8Array
	case "Uint16Array":
		return KindUint16Array
	case "Uint32Array":
		return KindUint32Array
	case "Float32Array":
		return KindFloat32Array
	case "Float64Array":
		return KindFloat64Array
	case "Map":
		return KindMap
	case "Set":
		return KindSet
	case "Date":
		return KindDate
	default:
		return KindUnknown
	}
}

const (
	flagStringMask = 1 << iota
	flagSymbolMask
	flagPrivateMask
	flagEnumOnly
	flagSetEnum
)

func (o Object) GetOwnPropertyNames() []string {
	var enumPtr *C.JSPropertyEnum
	var size C.uint32_t
	flags := C.int(flagStringMask | flagSymbolMask | flagPrivateMask)
	result := int(C.JS_GetOwnPropertyNames(o.context.raw, &enumPtr, &size, o.raw, flags))
	if result < 0 {
		return nil
	}
	enums := unsafe.Slice(enumPtr, size)
	properties := make([]string, size)
	for i := 0; i < int(size); i++ {
		enum := enums[i]
		properties[i] = atom{o.context, enum.atom}.String()
		C.JS_FreeAtom(o.context.raw, enum.atom)
	}
	C.js_free(o.context.raw, unsafe.Pointer(enumPtr))
	return properties
}

func (o Object) ToNative() any {
	switch o.Kind() {
	case KindPlainObject:
		names := o.GetOwnPropertyNames()
		retval := make(map[string]any, len(names))
		for _, name := range names {
			property, _ := o.GetProperty(name)
			retval[name] = property.ToNative()
		}
		return retval
	case KindArray:
		return o.Array().ToNative()
	case KindArrayBuffer:
		return o.ArrayBuffer().ToNative()
	case KindInt8Array:
		return TypedArray[int8]{o}.ToNative()
	case KindInt16Array:
		return TypedArray[int16]{o}.ToNative()
	case KindInt32Array:
		return TypedArray[int32]{o}.ToNative()
	case KindUint8Array:
		return TypedArray[uint8]{o}.ToNative()
	case KindUint16Array:
		return TypedArray[uint16]{o}.ToNative()
	case KindUint32Array:
		return TypedArray[uint32]{o}.ToNative()
	case KindFloat32Array:
		return TypedArray[float32]{o}.ToNative()
	case KindFloat64Array:
		return TypedArray[float64]{o}.ToNative()
	case KindMap:
		return o.Map().ToNative()
	case KindSet:
		return o.Set().ToNative()
	case KindDate:
		return o.Date().ToNative()
	default:
		return NotNative{}
	}
}

func (o Object) call(numArgs int, argsPtr *C.JSValue) C.JSValue {
	return C.JS_Call(o.context.raw, o.raw, C.JS_Null(), C.int(numArgs), argsPtr)
}

func (o Object) getProperty(name string) C.JSValue {
	return C.JS_GetPropertyStr(o.context.raw, o.raw, strPtr(name+"\x00"))
}

func (o Object) GetProperty(name string) (Value, error) {
	jsValue := o.getProperty(name)
	if err := o.context.checkException(jsValue); err != nil {
		return Value{}, err
	}
	C.JS_FreeValue(o.context.raw, jsValue)
	return Value{o.context, jsValue}, nil
}

func (o Object) setProperty(name string, value C.JSValue) {
	C.JS_SetPropertyStr(o.context.raw, o.raw, strPtr(name+"\x00"), value)
}

// []byte will be converted to Uint8Array since []byte and []uint8 is the same
func (o Object) SetProperty(name string, value any) {
	o.setProperty(name, o.context.toJsValue(value))
}

// Assume value is Object
func (v Value) Object() Object { return Object{v} }
