package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFunction(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		object := context.GlobalObject()
		counter := 0
		nativeFunc := func(call Call) (Value, error) {
			counter = call.NumArgs()
			retval := call.Arg(0).ToPrimitive().(int) + call.Arg(1).ToPrimitive().(int)
			return call.ToValue(retval), nil
		}
		object.SetFunc("test", nativeFunc)
		retval, err := context.Eval("test(1, 2);")
		assert.NoError(t, err)
		assert.Equal(t, 2, counter)
		assert.Equal(t, 3, retval.ToNative())
	})
}

func TestAddMethod(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value := context.ToValue(map[string]any{"a": 1})
		value.Object().SetFunc("test", func(call Call) (Value, error) {
			value, err := call.This().Object().GetProperty("a")
			if err != nil {
				return call.ToValue(Undefined), err
			}
			sum := value.ToPrimitive().(int) + call.Arg(0).ToPrimitive().(int)
			return call.ToValue(map[string]any{"sum": sum}), nil
		})
		context.GlobalObject().SetProperty("a", value)
		value, err := context.Eval(`a.test(2)`)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"sum": 3}, value.ToNative())
	})
}

type tuple struct{ l, r any }

func makeTuple(call Call) (Value, error) {
	value := tuple{call.Arg(0).ToPrimitive(), call.Arg(1).ToPrimitive()}
	return call.ToValue(value), nil
}

func TestAddConstructor(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		context.GlobalObject().SetFunc("Tuple", makeTuple, true)
		value, err := context.Eval(`new Tuple(1, "2")`)
		assert.NoError(t, err)
		assert.Equal(t, tuple{1, "2"}, value.ToNative())
	})
}

func BenchmarkCallGoFunction(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		nativeFunc := func(call Call) (Value, error) {
			retval := call.Arg(0).ToPrimitive().(int) + call.Arg(1).ToPrimitive().(int)
			return call.ToValue(retval), nil
		}
		global := context.GlobalObject()
		global.SetFunc("test", nativeFunc)
		retval, err := context.Eval("test(1, 2)")
		assert.NoError(b, err)
		assert.Equal(b, 3, retval.ToNative())
		for i := 0; i < b.N; i++ {
			_, _ = context.Eval("test(1, 2)")
		}
	})
}
