package quickjs

//#include "ffi.h"
import "C"

type Map struct{ Object }

func (m Map) Size() int {
	property, _ := m.GetProperty("size")
	return property.ToPrimitive().(int)
}

func (m Map) ToNative() map[any]any {
	arrayFn, _ := m.context.GlobalObject().GetProperty("Array")
	from, _ := arrayFn.Object().GetProperty("from")
	callValue := m.context.assert(from.Object().call(null, 1, &m.raw))
	array := Value{m.context, callValue}.Object().Array()
	length := array.Len()
	if length == 0 {
		return nil
	}
	retval := make(map[any]any, length)
	for i := 0; i < length; i++ {
		entry := array.Get(i).Object().Array()
		key := entry.Get(0).ToNative()
		retval[key] = entry.Get(1).ToNative()
	}
	C.JS_FreeValue(m.context.raw, callValue)
	return retval
}

// Assume object is Map
func (o Object) Map() Map { return Map{o} }
