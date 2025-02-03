#include <stdlib.h>
#include <string.h>
#include "libquickjs/quickjs.h"
#include "libquickjs/quickjs-libc.h"

static inline JS_BOOL JS_IsInt(JSValueConst val) { return JS_VALUE_GET_TAG(val) == JS_TAG_INT; }

static inline JSValue JS_Null() { return JS_NULL; }

static inline int JS_ValueTag(JSValueConst val) { return JS_VALUE_GET_TAG(val); }

static inline int JS_ValueRefCount(JSContext *ctx, JSValue v)
{
    if (JS_VALUE_HAS_REF_COUNT(v)) {
        JSRefCountHeader *p = (JSRefCountHeader *)JS_VALUE_GET_PTR(v);
        return p->ref_count;
    }
    return 0;
}

extern JSValue proxyCall(JSContext *ctx, JSValueConst this, int argc, JSValueConst *argv);
