package quickjs

//#include "ffi.h"
import "C"
import "reflect"

func (c *Context) toObject(value any) C.JSValue {
	var jsValue C.JSValue
	if callable, ok := value.(IndexCallable); ok {
		typeOf := reflect.ValueOf(value).Type()
		protoClass, ok := c.protoClasses[typeOf]
		if !ok {
			protoClass = C.JS_NewObject(c.raw)
			object := Value{c, protoClass}.Object()
			for i, name := range callable.MethodList() {
				object.setProperty(name, c.goIndexCall(i))
			}
			c.protoClasses[typeOf] = protoClass
		}
		jsValue = c.goObject(value, protoClass, c.runtime.goObject)
	} else {
		jsValue = c.goObject(value, c.goObjectProto, c.runtime.goObject)
	}
	return jsValue
}

// Structs will be wrapped into javascript object with specific prototype,
// if struct implements IndexCallable, return value from MethodList will
// be added to that object for calling go method from JS.
func (c *Context) ToObject(value any) Value {
	return Value{c, c.toObject(value)}
}
