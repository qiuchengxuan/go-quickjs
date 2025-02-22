package quickjs

//#include "ffi.h"
import "C"
import (
	"unsafe"
)

type goObjectData struct {
	value   any
	context *Context
}

//export goObjectFinalizer
func goObjectFinalizer(_ *C.JSRuntime, value C.JSValue) {
	dataPtr := C.JS_GetOpaque(value, C.JS_GetClassID(value))
	data := (*goObjectData)(dataPtr)
	delete(data.context.goValues, (uintptr)(C.JS_ValuePtr(value)))
	C.free(dataPtr)
}

type jsContext = C.JSContext
type jsValue = C.JSValue
type jsValueC = C.JSValueConst

type callback = func(*jsContext, jsValueC, jsValueC, []jsValueC, C.int) jsValueC

//export proxyCall
func proxyCall(ctx *jsContext, fn, this jsValueC, argc C.int, argv *jsValueC, flags C.int) jsValue {
	refs := unsafe.Slice(argv, argc)
	dataPtr := C.JS_GetOpaque(fn, C.JS_GetClassID(fn))
	data := (*goObjectData)(dataPtr)
	return data.value.(callback)(ctx, fn, this, refs, flags)
}
