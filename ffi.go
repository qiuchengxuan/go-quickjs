package quickjs

//#include "ffi.h"
import "C"
import (
	"unsafe"
)

type goObjectData struct {
	context *Context
	value   any
	flags   ObjectFlags
}

//export goObjectFinalizer
func goObjectFinalizer(_ *C.JSRuntime, value C.JSValue) {
	dataPtr := C.JS_GetOpaque(value, C.JS_GetClassID(value))
	data := (*goObjectData)(dataPtr)
	delete(data.context.goValues, (uintptr)(C.JS_ValuePtr(value)))
	C.free(dataPtr)
}

type jsCtx = C.JSContext
type jsVal = C.JSValue
type jsValCst = C.JSValueConst
type classID = C.JSClassID

func getObjectData(value C.JSValueConst) *goObjectData {
	return (*goObjectData)(C.JS_GetOpaque(value, C.JS_GetClassID(value)))
}

//export proxyCall
func proxyCall(_ *jsCtx, fn, this jsValCst, argc C.int, argv *jsValCst, flags C.int) jsVal {
	args := unsafe.Slice(argv, argc)
	data := getObjectData(fn)
	retval, err := data.value.(Func)(Call{data.context, fn, this, args, flags})
	if err != nil {
		return data.context.ThrowInternalError("%s", err)
	}
	return retval.raw
}

//export indexCall
func indexCall(_ *jsCtx, fn, this jsValCst, argc C.int, argv *jsValCst, flags C.int) jsVal {
	args := unsafe.Slice(argv, argc)
	data := getObjectData(this)
	index := int(uintptr(C.JS_GetOpaque(fn, C.JS_GetClassID(fn))))
	call := Call{data.context, fn, this, args, flags}
	retval, err := data.value.(IndexCallable).IndexCall(index, call)
	if err != nil {
		return data.context.ThrowInternalError("%s", err)
	}
	return retval.raw
}
