quickjs
=======

Go bindings to QuickJS: a fast, small, and embeddable ES2020 JavaScript interpreter.

Example
-------

```go
package main

import (
	"github.com/qiuchengxuan/go-quickjs"
)

func main() {
	quickjs.NewRuntime().NewContext().With(context *Context) {
		value, _ := context.Eval("new Map()")
		_ = value.ToNative().(map[string]any)

		context := quickjs.NewRuntime().NewContext()
		value, _ = context.Eval(`let value = "value"`)
		_ = context.GlobalObject().GetProperty("value").ToNative() // Should be "value"

		byteCode, _ := context.Compile("1 + 1")
		value, _ = context.EvalBinary(byteCode)
		value.ToNative() // Should be 2

		context.GlobalObject().SetProperty("value", 0.1)
		value, _ = context.Eval(`value + 0.1`)
		_ = value.ToNative() // should be 0.2

		counter := 0
		context.GlobalObject().SetProperty("sum", func(args ...any) any {
		    return args[0].(int) + args[1].(int)
		})
		value, _ = context.Eval("test(1, 2)")
		_ = value.ToNative() // should be 3 and counter should be 2
    })
}
```
