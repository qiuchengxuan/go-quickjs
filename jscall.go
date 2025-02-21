package quickjs

//#include "ffi.h"
import "C"

type Call struct {
	*Context
	fn, this C.JSValueConst
	args     []C.JSValueConst
}

func (c Call) NumArgs() int {
	return len(c.args)
}

func (c Call) Fn() Value {
	return Value{c.Context, c.fn}
}

func (c Call) This() Value {
	return Value{c.Context, c.this}
}

func (c Call) Arg(index int) Value {
	return Value{c.Context, c.args[index]}
}

type Func = func(...any) (any, error)
type RawFunc = func(call Call) (Value, error)

func (c *Context) naiveFunc(fn Func) C.JSValue {
	cb := func(_ *C.JSContext, _, _ C.JSValueConst, args []C.JSValueConst) C.JSValueConst {
		goArgs := make([]any, len(args))
		for i, arg := range args {
			goArgs[i] = Value{c, arg}.ToNative()
		}
		retval, err := fn(goArgs...)
		if err != nil {
			return c.ThrowInternalError("%s", err)
		}
		return c.toJsValue(retval)
	}
	return c.goObject(cb)
}

func (c *Context) rawFunc(rawFunc RawFunc) C.JSValue {
	cb := func(_ *C.JSContext, fn, this C.JSValueConst, args []C.JSValueConst) C.JSValueConst {
		value, err := rawFunc(Call{c, fn, this, args})
		if err != nil {
			return c.ThrowInternalError("%s", err)
		}
		return value.raw
	}
	return c.goObject(cb)
}
