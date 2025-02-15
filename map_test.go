package quickjs

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapToNative(t *testing.T) {
	globalConfig.ManualFree = true
	runtime := NewRuntime()
	guard := runtime.NewContext()
	guard.With(func(context *Context) {
		value, err := context.Eval(`new Map([["key", "value"], ["int", 1]])`)
		assert.NoError(t, err)
		assert.Equal(t, TypeObject, value.Type())
		expected := map[any]any{"key": "value", "int": 1}
		assert.Equal(t, expected, value.ToNative())
	})
	guard.Free()
	runtime.Free()
}

func TestMapFromNative(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		expected := map[any]any{1: 2, 3: 4}
		global := context.GlobalObject()
		global.SetProperty("whatever", expected)
		value, _ := global.GetProperty("whatever")
		assert.Equal(t, expected, value.ToNative())
	})
}

func BenchmarkMapToNative(b *testing.B) {
	globalConfig.ManualFree = true
	runtime := NewRuntime()
	guard := runtime.NewContext()
	guard.With(func(context *Context) {
		code := "new Map(Array.from(Array(16).keys()).map(v => [v.toString(), v]))"
		value, err := context.Eval(code)
		assert.NoError(b, err)
		expected := make(map[any]any, 16)
		for i := 0; i < 16; i++ {
			expected[strconv.Itoa(i)] = i
		}
		assert.Equal(b, expected, value.ToNative())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = value.ToNative()
		}
		b.StopTimer()
	})
	guard.Free()
	runtime.Free()
}

func BenchmarkMapFromNative(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		native := make(map[any]any, 16)
		for i := 0; i < 16; i++ {
			native[strconv.Itoa(i)] = i
		}
		global := context.GlobalObject()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			global.SetProperty("whatever", native)
		}
		b.StopTimer()
	})
}
