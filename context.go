package quickjs

//#include "ffi.h"
import "C"
import (
	"reflect"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type Context struct {
	runtime          *Runtime
	raw              *C.JSContext
	global           C.JSValue
	evalRet          C.JSValue
	goObjectProto    C.JSValueConst
	goFuncProto      C.JSValueConst
	goIndexCallProto C.JSValueConst
	goValues         map[uintptr]any
	objectKinds      map[C.JSValue]ObjectKind
	protoClasses     map[reflect.Type]C.JSValueConst
	free             atomic.Bool
}

func (c *Context) goObject(value any, proto jsValCst, classID classID, flags ObjectFlags) jsVal {
	jsObject := C.JS_NewObjectProtoClass(c.raw, proto, classID)
	c.goValues[(uintptr)(C.JS_ValuePtr(jsObject))] = value
	data := goObjectData{c, value, flags}
	dataPtr := C.malloc(C.size_t(unsafe.Sizeof(data)))
	*(*goObjectData)(dataPtr) = data
	C.JS_SetOpaque(jsObject, dataPtr)
	return jsObject
}

func (c *Context) goIndexCall(value int) C.JSValue {
	jsObject := C.JS_NewObjectProtoClass(c.raw, c.goIndexCallProto, c.runtime.goIndexCall)
	C.JS_SetOpaque(jsObject, unsafe.Pointer(uintptr(value)))
	return jsObject
}

func (c *Context) GlobalObject() GlobalObject {
	return GlobalObject{Value{c, c.global}.Object()}
}

func (c *Context) Compile(code string) (ByteCode, error) {
	codePtr := strPtr(code + "\x00")
	filename := "<input>\x00"
	flags := C.int(C.JS_EVAL_TYPE_GLOBAL | C.JS_EVAL_FLAG_COMPILE_ONLY)
	if C.JS_DetectModule(codePtr, strlen(code)) != 0 {
		flags |= C.JS_EVAL_TYPE_MODULE
	}
	jsValue := C.JS_Eval(c.raw, codePtr, strlen(code), strPtr(filename), flags)
	if err := c.checkException(jsValue); err != nil {
		return nil, err
	}
	var size C.size_t
	pointer := C.JS_WriteObject(c.raw, &size, jsValue, C.JS_WRITE_OBJ_BYTECODE)
	C.JS_FreeValue(c.raw, jsValue)
	if int(size) <= 0 {
		return nil, c.getException()
	}
	byteCode := C.GoBytes(unsafe.Pointer(pointer), C.int(size))
	C.js_free(c.raw, unsafe.Pointer(pointer))
	return byteCode, nil
}

func (c *Context) eval(code string) (C.JSValue, error) {
	codePtr := strPtr(code + "\x00")
	filename := "<input>\x00"
	flags := C.int(C.JS_EVAL_TYPE_GLOBAL)
	if C.JS_DetectModule(codePtr, strlen(code)) != 0 {
		flags |= C.JS_EVAL_TYPE_MODULE
	}
	jsValue := C.JS_Eval(c.raw, codePtr, strlen(code), strPtr(filename), flags)
	if err := c.checkException(jsValue); err != nil {
		return null, err
	}
	return jsValue, nil
}

// Return value must be consumed immediately before next Eval or EvalBinary
func (c *Context) Eval(code string) (Value, error) {
	C.JS_FreeValue(c.raw, c.evalRet)
	value, err := c.eval(code)
	c.evalRet = value
	return Value{c, value}, err
}

// Return value must be consumed immediately before next Eval or EvalBinary
func (c *Context) EvalBinary(byteCode ByteCode) (Value, error) {
	flags := C.int(C.JS_READ_OBJ_BYTECODE)
	object := C.JS_ReadObject(c.raw, bytesPtr(byteCode), C.size_t(len(byteCode)), flags)
	retval := c.assert(C.JS_EvalFunction(c.raw, c.assert(object)))
	C.JS_FreeValue(c.raw, c.evalRet)
	c.evalRet = retval
	return Value{c, retval}, nil
}

// Free context manually
func (c *Context) Free() {
	if c.free.Swap(true) {
		return
	}
	C.JS_FreeValue(c.raw, c.global)
	C.JS_FreeValue(c.raw, c.evalRet)
	for _, proto := range c.protoClasses {
		C.JS_FreeValue(c.raw, proto)
	}
	C.JS_FreeContext(c.raw)
	c.runtime.Free()
}

type ContextGuard struct{ context *Context }

// Manipulate Context with os thread locked
func (g ContextGuard) With(fn func(*Context)) {
	// Reason unknown, without locking os thread will cause quickjs throw strange exception
	runtime.LockOSThread()
	fn(g.context)
	runtime.UnlockOSThread()
}

// NOTE: unsafe
func (g ContextGuard) Unwrap() *Context { return g.context }

func (g ContextGuard) Free() { g.context.Free() }

func (r *Runtime) NewContext() ContextGuard {
	r.refCount.Add(1)
	C.js_std_init_handlers(r.raw)

	jsContext := C.JS_NewContext(r.raw)
	C.JS_AddIntrinsicBigFloat(jsContext)
	C.JS_AddIntrinsicBigDecimal(jsContext)
	C.JS_AddIntrinsicOperators(jsContext)
	C.JS_EnableBignumExt(jsContext, C.int(1))

	object := C.JS_GetGlobalObject(jsContext)
	context := &Context{runtime: r, raw: jsContext, global: object}
	context.goObjectProto = C.JS_NewObject(jsContext)
	C.JS_SetClassProto(jsContext, r.goObject, context.goObjectProto)
	context.goFuncProto = C.JS_NewObject(jsContext)
	C.JS_SetClassProto(jsContext, r.goFunc, context.goFuncProto)
	context.goIndexCallProto = C.JS_NewObject(jsContext)
	C.JS_SetClassProto(jsContext, r.goIndexCall, context.goIndexCallProto)
	objectKinds := make(map[C.JSValue]ObjectKind, KindMax)
	for i, name := range builtinKinds {
		jsValue, _ := context.GlobalObject().GetProperty(name)
		objectKinds[jsValue.raw] = ObjectKind(i + 1)
	}
	context.goValues = make(map[uintptr]any)
	context.objectKinds = objectKinds
	context.protoClasses = make(map[reflect.Type]C.JSValue)
	if !r.manualFree {
		runtime.SetFinalizer(context, func(c *Context) { c.Free() })
	}
	return ContextGuard{context}
}
