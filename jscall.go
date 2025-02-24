package quickjs

//#include "ffi.h"
import "C"

type Call struct {
	*Context
	fn, this C.JSValueConst
	args     []C.JSValueConst
	flags    C.int
}

func (c Call) NumArgs() int { return len(c.args) }

func (c Call) Fn() Value { return Value{c.Context, c.fn} }

func (c Call) This() Value { return Value{c.Context, c.this} }

func (c Call) Arg(index int) Value { return Value{c.Context, c.args[index]} }

func (c Call) Flags() uint { return uint(c.flags) }

type Func = func(call Call) (Value, error)

func (c *Context) rawFunc(rawFunc Func) C.JSValue {
	return c.goObject(rawFunc, c.goFuncProto, c.runtime.goFunc)
}
