package quickjs

type ByteCode []byte

type NotNative struct{ string }

func (n NotNative) String() string { return n.string }

type undefined struct{}

var Undefined *undefined = nil

type ViaJSON interface {
	ViaJSON()
}

type ProxyCaller interface {
	ProxyCall(Call) (Value, error)
}
