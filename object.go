package quickjs

//#include "ffi.h"
import "C"
import (
	"encoding/json"
	"unsafe"
)

type ObjectKind uint8

var builtinKinds = [17]string{
	"Object", "Boolean",
	"Number", "BigInt", "Date", "String",
	"Int8Array", "Int16Array", "Int32Array",
	"Uint8Array", "Uint16Array", "Uint32Array",
	"Float32Array", "Float64Array",
	"Map", "Set",
	"ArrayBuffer",
}

const (
	KindArray ObjectKind = iota
	KindPlainObject
	KindBoolean
	KindNumber
	KindBigInt
	KindDate
	KindString
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
	KindArrayBuffer
	KindUnknown
	KindMax = KindUnknown
)

type Object struct{ Value }

func (o Object) Kind() ObjectKind {
	if C.JS_IsArray(o.context.raw, o.raw) == 1 {
		return KindArray
	}
	property, _ := o.GetProperty("constructor")
	if kind, ok := o.context.objectKinds[property.raw]; ok {
		return kind
	}
	return KindUnknown
}

const (
	flagStringMask = 1 << iota
	flagSymbolMask
	flagPrivateMask
	flagEnumOnly
	flagSetEnum
)

func (o Object) HasProperty(name string) bool {
	atom := C.JS_NewAtom(o.context.raw, strPtr(name+"\x00"))
	retval := C.JS_HasProperty(o.context.raw, o.raw, atom) == 1
	C.JS_FreeAtom(o.context.raw, atom)
	return retval
}

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

func (o Object) plainObjectToNative() any {
	jsValue, _ := o.GetProperty("length")
	if length, ok := jsValue.ToPrimitive().(int); ok {
		retval := make([]any, length)
		for i := 0; i < length; i++ {
			jsValue := Value{o.context, o.getPropertyByIndex(uint32(i))}
			retval[i] = jsValue.ToNative()
		}
		return retval
	}
	names := o.GetOwnPropertyNames()
	retval := make(map[string]any, len(names))
	for _, name := range names {
		property, _ := o.GetProperty(name)
		retval[name] = property.ToNative()
	}
	return retval
}

func (o Object) ToNative() any {
	switch o.Kind() {
	case KindPlainObject:
		if classID := C.JS_GetClassID(o.raw); classID == o.context.runtime.goObject {
			data := getObjectData(o.raw)
			if data.flags&MapJSONFields > 0 {
				_ = o.JsonOut(data.value)
			}
			return data.value
		}
		return o.plainObjectToNative()
	case KindBoolean:
		return o.toBool()
	case KindNumber:
		return o.toNumber()
	case KindString:
		return o.String()
	case KindBigInt:
		return o.toBigInt()
	case KindDate:
		return o.Date().ToNative()
	case KindArray:
		return o.Array().ToNative()
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
	case KindArrayBuffer:
		return o.ArrayBuffer().ToNative()
	default:
		return NotNative{o.String()}
	}
}

func (o Object) JsonOut(out any) error {
	return json.Unmarshal([]byte(o.JSONify()), out)
}

func (o Object) call(this C.JSValue, numArgs int, argsPtr *C.JSValue) C.JSValue {
	return C.JS_Call(o.context.raw, o.raw, this, C.int(numArgs), argsPtr)
}

func (o Object) IsFunction() bool {
	return C.JS_IsFunction(o.context.raw, o.raw) == 1
}

func (o Object) Call(this Value, args ...any) (Value, error) {
	jsArgs := make([]C.JSValue, len(args))
	for i, arg := range args {
		jsArgs[i] = o.context.toValue(arg)
	}
	retval := o.call(this.raw, len(args), &jsArgs[0])
	if err := o.context.checkException(retval); err != nil {
		return o.context.ToValue(Undefined), err
	}
	return Value{o.context, retval}, nil
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

// A shortcut to o.SetValue(name, o.ToValue(value))
func (o Object) SetProperty(name string, value any) {
	o.setProperty(name, o.context.toValue(value))
}

func (o Object) SetValue(name string, value Value) {
	o.setProperty(name, value.raw)
}

func (o Object) getPropertyByIndex(index uint32) C.JSValue {
	return C.JS_GetPropertyUint32(o.context.raw, o.raw, C.uint32_t(index))
}

func (o Object) GetPropertyByIndex(index uint32) Value {
	jsValue := o.getPropertyByIndex(index)
	C.JS_FreeValue(o.context.raw, jsValue)
	return Value{o.context, jsValue}
}

func (o Object) setPropertyByIndex(index uint32, value C.JSValue) {
	C.JS_SetPropertyUint32(o.context.raw, o.raw, C.uint32_t(index), value)
}

func (o Object) SetPropertyByIndex(index uint32, value any) {
	o.setPropertyByIndex(index, o.context.toValue(value))
}

func (o Object) SetFunc(name string, fn Func) {
	o.setProperty(name, o.context.rawFunc(fn))
}

// Assume value is Object
func (v Value) Object() Object { return Object{v} }

type GlobalObject struct{ Object }

func (o GlobalObject) SetFunc(name string, fn Func, constructor ...bool) {
	jsValue := o.context.rawFunc(fn)
	if len(constructor) > 0 && constructor[0] {
		C.JS_SetConstructorBit(o.context.raw, jsValue, 1)
	}
	o.setProperty(name, jsValue)
}
