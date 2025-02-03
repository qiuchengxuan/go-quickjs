package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("[1, 2]")
		assert.NoError(t, err)
		assert.Equal(t, []any{1, 2}, value.ToNative())
	})
}

func TestArrayBuffer(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("Uint8Array.from([1, 2]).buffer")
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2}, value.ToNative())
	})
}

func TestFromTypedArray(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("Uint16Array.from([1, 2])")
		assert.NoError(t, err)
		assert.Equal(t, []uint16{1, 2}, value.ToNative())
		value, err = context.Eval("Int32Array.from([-1, -2])")
		assert.NoError(t, err)
		assert.Equal(t, []int32{-1, -2}, value.ToNative())
		value, err = context.Eval("Float64Array.from([1.2, -2.1])")
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.2, -2.1}, value.ToNative())
	})
}
