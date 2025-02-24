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

//export proxyCall
func proxyCall(_ *jsContext, fn, this jsValueC, argc C.int, argv *jsValueC, flags C.int) jsValue {
	args := unsafe.Slice(argv, argc)
	dataPtr := C.JS_GetOpaque(fn, C.JS_GetClassID(fn))
	data := (*goObjectData)(dataPtr)
	retval, err := data.value.(Func)(Call{data.context, fn, this, args, flags})
	if err != nil {
		return data.context.ThrowInternalError("%s", err)
	}
	return retval.raw
}

//export indexCall
func indexCall(_ *jsContext, fn, this jsValueC, argc C.int, argv *jsValueC, flags C.int) jsValue {
	args := unsafe.Slice(argv, argc)
	dataPtr := C.JS_GetOpaque(this, C.JS_GetClassID(this))
	data := (*goObjectData)(dataPtr)
	index := int(uintptr(C.JS_GetOpaque(fn, C.JS_GetClassID(fn))))
	call := Call{data.context, fn, this, args, flags}
	retval, err := data.value.(IndexCallable).IndexCall(index, call)
	if err != nil {
		return data.context.ThrowInternalError("%s", err)
	}
	return retval.raw
}
