package quickjs

//#include "ffi.h"
import "C"
import (
	"bytes"
	"encoding/json"
	"reflect"
	"unsafe"
)

type ObjectFlags uint

const (
	// Able to get or set struct fields with json tag
	MapJSONFields ObjectFlags = 1 << iota
)

func (c *Context) toObject(value any, flags ObjectFlags) C.JSValue {
	var jsValue C.JSValue
	if callable, ok := value.(IndexCallable); ok {
		typeOf := reflect.ValueOf(value).Type()
		protoClass, ok := c.protoClasses[typeOf]
		if !ok {
			protoClass = C.JS_NewObject(c.raw)
			object := Value{c, protoClass}.Object()
			for i, name := range callable.Methods() {
				object.setProperty(name, c.goIndexCall(i))
			}
			c.protoClasses[typeOf] = protoClass
		}
		jsValue = c.goObject(value, protoClass, c.runtime.goObject, flags)
	} else {
		jsValue = c.goObject(value, c.goObjectProto, c.runtime.goObject, flags)
	}
	if flags&MapJSONFields > 0 {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(value); err != nil {
			return null
		}
		buf.WriteByte(0)
		data := buf.Bytes()
		dataPtr := (*C.char)(unsafe.Pointer(&data[0]))
		source := Value{c, C.JS_ParseJSON(c.raw, dataPtr, sliceSize(data)-1, nil)}.Object()
		target := Value{c, jsValue}.Object()
		for _, name := range source.GetOwnPropertyNames() {
			target.setProperty(name, source.getProperty(name))
		}
		C.JS_FreeValue(c.raw, source.raw)
	}
	return jsValue
}

// Structs will be wrapped into javascript object with specific prototype,
// if struct implements IndexCallable, return value from Methods will
// be added to that object for calling go method from JS.
func (c *Context) ToObject(value any, options ...ObjectFlags) Value {
	var flags ObjectFlags
	for _, flag := range options {
		flags |= flag
	}
	return Value{c, c.toObject(value, flags)}
}
