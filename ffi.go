package quickjs

//#include "ffi.h"
import "C"
import (
	"unsafe"
)

type interfaceData struct {
	value  any
	id     uint32
	values map[uint32]any
}

//export goInterfaceFinalizer
func goInterfaceFinalizer(_ *C.JSRuntime, value C.JSValue) {
	dataPtr := C.JS_GetOpaque(value, C.JS_GetClassID(value))
	data := (*interfaceData)(dataPtr)
	delete(data.values, data.id)
	C.free(dataPtr)
}

type callback = func(*C.JSContext, C.JSValueConst, []C.JSValueConst) C.JSValueConst

//export proxyCall
func proxyCall(ctx *C.JSContext, fn, this C.JSValueConst, argc C.int, argv *C.JSValueConst, flags C.int) C.JSValue {
	refs := unsafe.Slice(argv, argc)
	dataPtr := C.JS_GetOpaque(fn, C.JS_GetClassID(fn))
	value := (*interfaceData)(dataPtr).value
	return value.(callback)(ctx, this, refs)
}
