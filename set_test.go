package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetToNative(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval(`new Set([1, 2, 1])`)
		assert.NoError(t, err)
		assert.Equal(t, []any{1, 2}, value.ToNative())
	})
}

func BenchmarkSetToNative(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval("new Set(Array.from(Array(16).keys()))")
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
