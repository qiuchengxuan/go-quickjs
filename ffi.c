#include "libquickjs/quickjs.h"
#include "_cgo_export.h"

JSClassDef go_object_class = {
    "goObject",
    .finalizer = goObjectFinalizer,
    .call = proxyCall,
};

JSValue ThrowTypeError(JSContext *ctx, const char *fmt) { return JS_ThrowTypeError(ctx, "%s", fmt); }
