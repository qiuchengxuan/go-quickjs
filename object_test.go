package quickjs

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainObject(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		retval, err := context.Eval("new Object([1, 2])")
		assert.NoError(t, err)
		expected := []any{1, 2}
		assert.Equal(t, expected, retval.ToNative())

		context.GlobalObject().SetProperty("plain", expected)
		retval, err = context.GlobalObject().GetProperty("plain")
		assert.NoError(t, err)
		assert.Equal(t, expected, retval.ToNative())
	})
	NewRuntime().NewContext().With(func(context *Context) {
		retval, err := context.Eval("new Object({a: 1, b: {c: 2}})")
		assert.NoError(t, err)
		expected := map[string]any{"a": 1, "b": map[string]any{"c": 2}}
		assert.Equal(t, expected, retval.ToNative())

		context.GlobalObject().SetProperty("plain2", expected)
		retval, err = context.GlobalObject().GetProperty("plain2")
		assert.NoError(t, err)
		assert.Equal(t, expected, retval.ToNative())
	})
}

func TestSetProperty(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		object := context.GlobalObject()
		for _, value := range []any{1, "1", 0.1, true} {
			object.SetProperty("value", value)
			property, err := object.GetProperty("value")
			assert.NoError(t, err)
			assert.Equal(t, value, property.ToNative())
		}
		values := []any{
			int8(1), int16(1), int32(1), int64(1),
			uint(1), uint8(1), uint16(1), uint32(1), uint64(1),
		}
		for _, value := range values {
			object.SetProperty("value", value)
			var expected int
			valueOf := reflect.ValueOf(value)
			if valueOf.CanInt() {
				expected = int(valueOf.Int())
			} else {
				expected = int(valueOf.Uint())
			}
			property, err := object.GetProperty("value")
			assert.NoError(t, err)
			assert.Equal(t, expected, property.ToNative())
		}
		values = []any{
			uint(math.MaxUint), uint32(math.MaxUint32),
			int64(math.MaxInt64), uint64(math.MaxUint64),
		}
		for _, value := range values {
			object.SetProperty("value", value)
			var expected float64
			valueOf := reflect.ValueOf(value)
			if valueOf.CanInt() {
				expected = float64(valueOf.Int())
			} else {
				expected = float64(valueOf.Uint())
			}
			property, err := object.GetProperty("value")
			assert.NoError(t, err)
			assert.Equal(t, expected, property.ToNative())
		}
		values = []any{
			[]int8{1}, []int16{1}, []uint16{1}, []int32{1}, []uint32{1},
			[]float32{1}, []float64{1},
		}
		for _, value := range values {
			object.SetProperty("value", value)
			property, err := object.GetProperty("value")
			assert.NoError(t, err)
			assert.Equal(t, value, property.ToNative())
		}
	})
}

func TestFinalizer(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		proto, classID := context.goObjectProto, context.runtime.goObject
		Value{context, context.goObject(nil, proto, classID)}.free()
		assert.Zero(t, 0, len(context.goValues))
	})
}

type jsonMarshaller struct {
	A int    `json:"a"`
	B string `json:"b"`
}

func (m *jsonMarshaller) MarshalJSON() ([]byte, error) {
	type helper jsonMarshaller
	return json.Marshal((*helper)(m))
}

func TestJsonMarshal(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		value, err := context.Eval(`new Object({a: 1, b: "2"})`)
		assert.NoError(t, err)
		expected := `{"a":1,"b":"2"}`
		assert.Equal(t, expected, value.JSONify())

		var actual jsonMarshaller
		assert.NoError(t, value.Object().JsonOut(&actual))
		assert.Equal(t, jsonMarshaller{1, "2"}, actual)

		global := context.GlobalObject()
		global.SetProperty("jsonValue", &jsonMarshaller{1, "2"})
		value, err = global.GetProperty("jsonValue")
		assert.NoError(t, err)
		assert.Equal(t, expected, value.JSONify())
	})
}

func TestNativeCall(t *testing.T) {
	NewRuntime().NewContext().With(func(context *Context) {
		object, _ := context.GlobalObject().GetProperty("Object")
		this := context.ToValue(nil)
		value, err := object.Object().Call(this, "test")
		assert.NoError(t, err)
		assert.Equal(t, "test", value.ToNative())
	})
}

func BenchmarkGetKind(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		retval, err := context.Eval("new Date()")
		assert.NoError(b, err)
		assert.Equal(b, KindDate, retval.Object().Kind())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = retval.Object().Kind()
		}
	})
}

func BenchmarkObjectFromNative(b *testing.B) {
	NewRuntime().NewContext().With(func(context *Context) {
		expected := make(map[string]any)
		for i := 0; i < 16; i++ {
			expected[strconv.Itoa(i)] = i
		}
		global := context.GlobalObject()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			global.SetProperty("whatever", expected)
		}
		b.StopTimer()
	})
}
