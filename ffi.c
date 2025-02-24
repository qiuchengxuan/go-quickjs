#include "libquickjs/quickjs.h"
#include "_cgo_export.h"

JSClassDef go_object_class = {
    "goObject",
    .finalizer = goObjectFinalizer,
};

JSClassDef go_function_class = {
    "goFunction",
    .finalizer = goObjectFinalizer,
    .call = proxyCall,
};

JSClassDef go_indexcall_class = {
    "goIndexCall",
    .call = indexCall,
};

JSValue ThrowInternalError(JSContext *ctx, const char *fmt) {
    return JS_ThrowInternalError(ctx, "%s", fmt);
}
