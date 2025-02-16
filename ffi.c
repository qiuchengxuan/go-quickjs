#include "libquickjs/quickjs.h"
#include "_cgo_export.h"

JSClassDef go_function_class = {
    "goFunction",
    .finalizer = goInterfaceFinalizer,
    .call = proxyCall,
};
