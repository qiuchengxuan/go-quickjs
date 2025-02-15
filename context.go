package quickjs

//#include "ffi.h"
import "C"
import (
	"math"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type Context struct {
	runtime     *Runtime
	raw         *C.JSContext
	filename    C.int
	global      C.JSValue
	proxy       C.JSValue
	makeProxy   C.JSValue
	evalRet     C.JSValue
	goValues    map[any]C.JSValue
	objectKinds map[C.JSValue]ObjectKind
	free        atomic.Bool
}

func (c *Context) getException() error {
	value := Value{c, C.JS_GetException(c.raw)}
	cause := value.String()
	stack, _ := value.Object().GetProperty("stack")
	if stack.Type() == TypeUndefined {
		return &Error{Cause: cause}
	}
	err := &Error{Cause: cause, Stack: stack.String()}
	C.JS_FreeValue(c.raw, value.raw)
	return err
}

func (c *Context) checkException(value C.JSValue) error {
	if C.JS_IsException(value) == 1 {
		return c.getException()
	}
	return nil
}

func (c *Context) assert(value C.JSValue) C.JSValue {
	if err := c.checkException(value); err != nil {
		panic(err)
	}
	return value
}

func (c *Context) addCallback(callback *callback) C.JSValue {
	pointer := math.Float64frombits(uint64(uintptr(unsafe.Pointer(callback))))
	handler := C.JS_NewFloat64(c.raw, C.double(pointer))
	args := []C.JSValue{c.proxy, handler}
	return C.JS_Call(c.raw, c.makeProxy, null, C.int(len(args)), &args[0])
}

func (c *Context) addNaiveFunc(fn NaiveFunc) C.JSValue {
	callback := func(_ *C.JSContext, _ C.JSValueConst, args []C.JSValueConst) C.JSValueConst {
		goArgs := make([]any, len(args))
		for i, arg := range args {
			goArgs[i] = Value{c, arg}.ToNative()
		}
		return c.toJsValue(fn(goArgs...))
	}
	callbackPtr := &callback
	jsValue := c.addCallback(callbackPtr)
	C.JS_DupValue(c.raw, jsValue)
	c.goValues[callbackPtr] = jsValue
	return jsValue
}

func (c *Context) GlobalObject() Object {
	return Value{c, c.global}.Object()
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
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
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

func (c *Context) RunGC() {
	c.runtime.RunGC()
	for value, jsValue := range c.goValues {
		if C.JS_ValueRefCount(c.raw, jsValue) <= 1 {
			delete(c.goValues, value)
		}
	}
}

// Free context manually
func (c *Context) Free() {
	if c.free.Swap(true) {
		return
	}
	for _, jsValue := range c.goValues {
		C.JS_FreeValue(c.raw, jsValue)
	}
	toFree := []C.JSValue{c.global, c.proxy, c.makeProxy, c.evalRet}
	for _, jsValue := range toFree {
		C.JS_FreeValue(c.raw, jsValue)
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
	g.context.RunGC()
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

	fn := (*C.JSCFunction)(unsafe.Pointer(C.proxyCall))
	object := C.JS_GetGlobalObject(jsContext)
	proxy := C.JS_NewCFunction(jsContext, fn, nil, C.int(0))
	context := &Context{
		runtime: r, raw: jsContext, global: object, proxy: proxy,
		goValues: make(map[any]C.JSValue),
	}
	objectKinds := make(map[C.JSValue]ObjectKind, KindDate+1)
	for i, name := range builtinKinds {
		jsValue, _ := context.GlobalObject().GetProperty(name)
		objectKinds[jsValue.raw] = ObjectKind(i + 1)
	}
	context.objectKinds = objectKinds
	if !globalConfig.ManualFree {
		runtime.SetFinalizer(context, func(c *Context) { c.Free() })
	}
	makeProxy := "(proxy, handler) => function() { return proxy.call(handler, ...arguments) }"
	context.makeProxy, _ = context.eval(makeProxy)
	return ContextGuard{context}
}
