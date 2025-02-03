package quickjs

//#include "ffi.h"
import "C"

type atom struct {
	context *Context
	raw     C.JSAtom
}

func (a atom) String() string {
	cStr := C.JS_AtomToCString(a.context.raw, a.raw)
	retval := C.GoString(cStr)
	C.JS_FreeCString(a.context.raw, cStr)
	return retval
}
