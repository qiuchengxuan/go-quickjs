package quickjs

type ByteCode []byte

type NotNative struct{ string }

func (n NotNative) String() string { return n.string }

type undefined struct{}

var Undefined *undefined = nil

type QuickjsJsonMarshal interface {
	QuickjsJsonMarshal()
}

type IndexCallable interface {
	// List of method names to be added as method
	MethodList() []string
	// Index is the corresponding method list index
	IndexCall(int, Call) (Value, error)
}
