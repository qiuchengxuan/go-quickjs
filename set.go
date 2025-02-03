package quickjs

//#include "ffi.h"
import "C"

type Set struct{ Object }

func (s Set) Size() int {
	property, _ := s.GetProperty("size")
	return property.ToPrimitive().(int)
}

func (s Set) ToNative() []any {
	arrayFn, _ := s.context.GlobalObject().GetProperty("Array")
	from, _ := arrayFn.Object().GetProperty("from")
	callValue := s.context.assert(from.Object().call(1, &s.raw))
	array := Value{s.context, callValue}.Object().Array()
	length := array.Len()
	retval := make([]any, length)
	for i := 0; i < length; i++ {
		retval[i] = array.Get(i).ToNative()
	}
	C.JS_FreeValue(s.context.raw, callValue)
	return retval
}

// Assume object is Set
func (o Object) Set() Set { return Set{o} }
