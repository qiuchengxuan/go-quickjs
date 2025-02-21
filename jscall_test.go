package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFunction(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		object := context.GlobalObject()
		counter := 0
		naiveFunc := func(args ...any) (any, error) {
			counter = len(args)
			return args[0].(int) + args[1].(int), nil
		}
		object.AddFunc("test", naiveFunc)
		retval, err := context.Eval("test(1, 2);")
		assert.NoError(t, err)
		assert.Equal(t, 2, counter)
		assert.Equal(t, 3, retval.ToNative())
	})
}

func TestAddMethod(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value := context.ToValue(map[string]any{"a": 1})
		value.Object().AddFunc("test", func(call Call) (Value, error) {
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
