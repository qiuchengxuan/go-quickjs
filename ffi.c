#include "_cgo_export.h"
#include "libquickjs/quickjs.h"

JSClassDef go_classes[3] = {{
    "goObject",
    .finalizer = goObjectFinalizer,
}, {
    "goFunction",
    .finalizer = goObjectFinalizer,
    .call = proxyCall,
}, {
    "goIndexCall",
    .call = indexCall,
}};

JSValue ThrowInternalError(JSContext *ctx, const char *fmt) {
    return JS_ThrowInternalError(ctx, "%s", fmt);
}
