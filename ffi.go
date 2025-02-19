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

type callback = func(*C.JSContext, C.JSValueConst, []C.JSValueConst) C.JSValueConst

//export proxyCall
func proxyCall(
	ctx *C.JSContext, fn, this C.JSValueConst, argc C.int, argv *C.JSValueConst, flags C.int,
) C.JSValue {
	refs := unsafe.Slice(argv, argc)
	dataPtr := C.JS_GetOpaque(fn, C.JS_GetClassID(fn))
	data := (*goObjectData)(dataPtr)
	return data.value.(callback)(ctx, this, refs)
}
