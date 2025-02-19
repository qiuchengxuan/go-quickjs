package quickjs

import "encoding/json"

type ByteCode []byte

type NotNative struct{}

type NaiveFunc = func(...any) any

type JSONValue interface {
	json.Marshaler
	json.Unmarshaler
}

type AsJSONValue[T any] struct{ value T }

func (c AsJSONValue[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c AsJSONValue[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.value)
}
