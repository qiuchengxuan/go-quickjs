package quickjs

type ByteCode []byte

type NotNative struct{}

type undefined struct{}

var Undefined *undefined = nil

type NaiveFunc = func(...any) (any, error)
