package quickjs

//#include "ffi.h"
import "C"
import (
	"runtime"
	"sync/atomic"
	"unsafe"

	_ "github.com/qiuchengxuan/go-quickjs/libquickjs"
)

type MemoryUsage struct {
	MallocSize, MallocLimit, MemoryUsedSize int64
	MallocCount                             int64
	MemoryUsedCount                         int64
	AtomCount, AtomSize                     int64
	StrCount, StrSize                       int64
	ObjCount, ObjSize                       int64
	PropCount, PropSize                     int64
	ShapeCount, ShapeSize                   int64
	JsFuncCount, JsFuncSize, JsFuncCodeSize int64
	JsFuncPc2lineCount, JsFuncPc2lineSize   int64
	CFuncCount, ArrayCount                  int64
	FastArrayCount, FastArrayElements       int64
	BinaryObjectCount, BinaryObjectSize     int64
}

type Error struct {
	Cause, Stack string
}

func (e Error) Error() string {
	return e.Cause
}

type Runtime struct {
	raw        *C.JSRuntime
	manualFree bool
	goObject   C.JSClassID
	goFunc     C.JSClassID
	refCount   atomic.Int32
}

func (r *Runtime) GetMemoryUsage() MemoryUsage {
	var usage MemoryUsage
	C.JS_ComputeMemoryUsage(r.raw, (*C.JSMemoryUsage)(unsafe.Pointer(&usage)))
	return usage
}

func (r *Runtime) RunGC() {
	C.JS_RunGC(r.raw)
}

// Free runtime manually
func (r *Runtime) Free() {
	if r.refCount.Add(-1) == 0 {
		C.JS_FreeRuntime(r.raw)
	}
}

func NewRuntime(config ...Config) *Runtime {
	jsRuntime := &Runtime{raw: C.JS_NewRuntime()}
	if len(config) > 0 {
		config := config[0]
		if size := config.MaxStackSize; size >= 0 {
			C.JS_SetMaxStackSize(jsRuntime.raw, C.size_t(size))
		}
		jsRuntime.manualFree = config.ManualFree
	}
	C.JS_NewClassID(&jsRuntime.goObject)
	C.JS_NewClass(jsRuntime.raw, jsRuntime.goObject, &C.go_object_class)
	C.JS_NewClassID(&jsRuntime.goFunc)
	C.JS_NewClass(jsRuntime.raw, jsRuntime.goFunc, &C.go_function_class)
	jsRuntime.refCount.Add(1)
	if !jsRuntime.manualFree {
		runtime.SetFinalizer(jsRuntime, func(r *Runtime) { r.Free() })
	}
	return jsRuntime
}
