package quickjs

//#include "ffi.h"
import "C"
import (
	"math"
)

func (c *Context) toJsValue(value any) C.JSValue {
	switch value := value.(type) {
	case bool:
		intValue := 0
		if value {
			intValue = 1
		}
		return C.JS_NewBool(c.raw, C.int(intValue))
	case int8:
		return C.JS_NewInt32(c.raw, C.int32_t(value))
	case int16:
		return C.JS_NewInt32(c.raw, C.int32_t(value))
	case int32:
		return C.JS_NewInt32(c.raw, C.int32_t(value))
	case int64:
		return C.JS_NewInt64(c.raw, C.int64_t(value))
	case int:
		return C.JS_NewInt64(c.raw, C.int64_t(value))
	case uint8:
		return C.JS_NewInt32(c.raw, C.int32_t(value))
	case uint16:
		return C.JS_NewInt32(c.raw, C.int32_t(value))
	case uint32:
		return C.JS_NewInt64(c.raw, C.int64_t(value))
	case uint64:
		if value <= math.MaxInt64 {
			return C.JS_NewInt64(c.raw, C.int64_t(value))
		}
		return C.JS_NewFloat64(c.raw, C.double(value))
	case uint:
		if value <= math.MaxInt64 {
			return C.JS_NewInt64(c.raw, C.int64_t(value))
		}
		return C.JS_NewFloat64(c.raw, C.double(value))
	case float32:
		return C.JS_NewFloat64(c.raw, C.double(value))
	case float64:
		return C.JS_NewFloat64(c.raw, C.double(value))
	case string:
		newStr := value + "\x00"
		return C.JS_NewString(c.raw, strPtr(newStr))
	case []int8:
		return newTypedArray(c, value, typedArrayInt8)
	case []uint8: // also []byte
		return newTypedArray(c, value, typedArrayUint8)
	case []int16:
		return newTypedArray(c, value, typedArrayInt16)
	case []uint16:
		return newTypedArray(c, value, typedArrayUint16)
	case []int32:
		return newTypedArray(c, value, typedArrayInt32)
	case []uint32:
		return newTypedArray(c, value, typedArrayUint32)
	case []float32:
		return newTypedArray(c, value, typedArrayFloat32)
	case []float64:
		return newTypedArray(c, value, typedArrayFloat64)
	case map[string]any:
		object := Value{c, C.JS_NewObject(c.raw)}.Object()
		for key, value := range value {
			object.setProperty(key, c.toJsValue(value))
		}
		return object.raw
	case NaiveFunc:
		return c.addNaiveFunc(value)
	default:
		return C.JS_Null()
	}
}
