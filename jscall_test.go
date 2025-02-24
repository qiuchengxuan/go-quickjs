package quickjs

import (
	"math/big"
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

type bigInt struct{ big.Int }

func (b *bigInt) MethodList() []string {
	methods := [3]string{"add", "sub", "nop"}
	return methods[:]
}

func (b *bigInt) IndexCall(index int, call Call) (Value, error) {
	switch index {
	case 0:
		rhs := call.Arg(0).ToNative()
		b.Add(&b.Int, big.NewInt(int64(rhs.(int))))
		return call.ToValue(b), nil
	case 1:
		rhs := call.Arg(0).ToNative()
		b.Sub(&b.Int, big.NewInt(int64(rhs.(int))))
		return call.ToValue(b), nil
	case 2:
		return call.ToValue(nil), nil
	default:
		panic("unreachable")
	}
}

func TestIndexCall(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		context.GlobalObject().SetProperty("ut", &bigInt{*big.NewInt(100)})
		value, err := context.Eval(`ut.add(2n)`)
		assert.NoError(t, err)
		assert.Equal(t, &bigInt{*big.NewInt(102)}, value.ToNative())

		value, err = context.Eval(`ut.sub(2n)`)
		assert.NoError(t, err)
		assert.Equal(t, &bigInt{*big.NewInt(100)}, value.ToNative())
	})
}

func BenchmarkNativeCall(b *testing.B) {
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

func BenchmarkIndexCall(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		context.GlobalObject().SetProperty("ut", &bigInt{*big.NewInt(0)})
		for i := 0; i < b.N; i++ {
			_, _ = context.Eval("ut.nop(1n)")
		}
	})
}
