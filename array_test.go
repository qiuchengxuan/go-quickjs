package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayToNative(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("[1, 2]")
		assert.NoError(t, err)
		assert.Equal(t, []any{1, 2}, value.ToNative())
	})
}

func TestArrayBufferToNative(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("Uint8Array.from([1, 2]).buffer")
		assert.NoError(t, err)
		assert.Equal(t, []byte{1, 2}, value.ToNative())
	})
}

func TestTypedArrayToNative(t *testing.T) {
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

func BenchmarkArrayToNative(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("Array.from(Array(16).keys())")
		assert.NoError(b, err)
		expected := make([]any, 16)
		for i := 0; i < 16; i++ {
			expected[i] = i
		}
		assert.Equal(b, expected, value.ToNative())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = value.ToNative()
		}
		b.StopTimer()
	})
}

func BenchmarkArrayFromNative(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		expected := make([]any, 16)
		for i := 0; i < 16; i++ {
			expected[i] = i
		}
		global := context.GlobalObject()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			global.SetProperty("whatever", expected)
		}
		b.StopTimer()
	})
}
