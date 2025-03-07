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
	refCount   atomic.Int32

	goObject, goFunc, goIndexCall C.JSClassID
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
	classIDs := [3]*C.JSClassID{&jsRuntime.goObject, &jsRuntime.goFunc, &jsRuntime.goIndexCall}
	for i, classID := range classIDs {
		C.JS_NewClassID(classID)
		C.JS_NewClass(jsRuntime.raw, *classID, &C.go_classes[i])
	}
	jsRuntime.refCount.Add(1)
	if !jsRuntime.manualFree {
		runtime.SetFinalizer(jsRuntime, func(r *Runtime) { r.Free() })
	}
	return jsRuntime
}
