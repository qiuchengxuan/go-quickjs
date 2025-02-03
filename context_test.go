package quickjs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManualFree(t *testing.T) {
	globalConfig.ManualFree = true
	runtime := NewRuntime()
	context := runtime.NewContext()
	context.Free()
	runtime.Free()

	runtime = NewRuntime()
	context = runtime.NewContext()
	runtime.Free()
	context.Free()
}

func TestCompile(t *testing.T) {
	globalConfig.ManualFree = true
	runtime := NewRuntime()
	guard := runtime.NewContext()

	guard.With(func(context *Context) {
		byteCode, err := context.Compile("1 + 1")
		assert.NoError(t, err)
		retval, err := context.EvalBinary(byteCode)
		assert.NoError(t, err)
		assert.Equal(t, 2, retval.ToNative())
	})

	guard.Free()
	runtime.Free()
}
