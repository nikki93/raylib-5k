#include "core/core.hh"


//
// JS interface
//

// Elements

JS_DEFINE(void, JS_uiElemOpenStart, (const char *tag),
    { IncrementalDOM.elementOpenStart(UTF8ToString(tag)); })

JS_DEFINE(void, JS_uiElemOpenStartKeyInt, (const char *tag, int key),
    { IncrementalDOM.elementOpenStart(UTF8ToString(tag), key); })

JS_DEFINE(void, JS_uiElemOpenStartKeyString, (const char *tag, const char *key),
    { IncrementalDOM.elementOpenStart(UTF8ToString(tag), UTF8ToString(key)); })

JS_DEFINE(void, JS_uiElemOpenEnd, (), { IncrementalDOM.elementOpenEnd(); })

JS_DEFINE(
    void, JS_uiElemClose, (const char *tag), { IncrementalDOM.elementClose(UTF8ToString(tag)); })

JS_DEFINE(int, JS_uiGetToken, (), {
  const target = IncrementalDOM.currentElement();
  let token = target.__UIToken;
  if (!token) {
    token = UI.nextToken++;
    target.__UIToken = token;
  }
  return token;
})


// Attributes

JS_DEFINE(void, JS_uiAttrInt, (const char *name, int value),
    { IncrementalDOM.attr(UTF8ToString(name), value); })

JS_DEFINE(void, JS_uiAttrFloat, (const char *name, float value),
    { IncrementalDOM.attr(UTF8ToString(name), value); })

JS_DEFINE(void, JS_uiAttrDouble, (const char *name, double value),
    { IncrementalDOM.attr(UTF8ToString(name), value); })

JS_DEFINE(void, JS_uiAttrString, (const char *name, const char *value),
    { IncrementalDOM.attr(UTF8ToString(name), UTF8ToString(value)); })

JS_DEFINE(
    void, JS_uiAttrEmpty, (const char *name), { IncrementalDOM.attr(UTF8ToString(name), ""); })

JS_DEFINE(void, JS_uiAttrClass, (const char *value),
    { IncrementalDOM.attr("class", UTF8ToString(value)); })


// Text

JS_DEFINE(void, JS_uiText, (const char *value), { IncrementalDOM.text(UTF8ToString(value)); })


// Events

JS_DEFINE(int, JS_uiGetEventCount, (const char *type), {
  const typeStr = UTF8ToString(type);
  const target = IncrementalDOM.currentElement();
  if (!target.__UIIsEventListenerSetup) {
    UI.setupEventListener(target, typeStr);
    target.__UIIsEventListenerSetup = true;
  }
  const counts = window.UI.eventCounts.get(target);
  if (typeof(counts) == "undefined") {
    return 0;
  }
  const count = counts[typeStr];
  if (typeof(count) == "undefined") {
    return 0;
  }
  delete counts[typeStr];
  return count;
})

JS_DEFINE(void, JS_uiClearEventCounts, (), { window.UI.eventCounts = new WeakMap(); })

JS_DEFINE(bool, JS_uiIsKeyboardCaptured, (), { return UI.keyboardCaptureCount > 0; })


// Values

JS_DEFINE(char *, JS_uiGetValue, (), {
  let value = IncrementalDOM.currentElement().value;
  if (!value) {
    value = "";
  }
  return allocate(intArrayFromString(value), ALLOC_NORMAL);
});

JS_DEFINE(void, JS_uiSetValue, (const char *value),
    { IncrementalDOM.currentElement().value = UTF8ToString(value); });


// Patch

JS_EXPORT void JS_uiCallPatchFunc() {
  if (uiPatchState.wrapper) {
    uiPatchState.wrapper();
    uiPatchState = {};
  }
}

JS_DEFINE(void, JS_uiPatch, (const char *id), {
  const el = document.getElementById(UTF8ToString(id));
  if (el) {
    IncrementalDOM.patch(el, Module._JS_uiCallPatchFunc.bind(Module));
  }
})
