#pragma once
#include "core/core.hh"


//
// JS interface
//

// Elements
extern "C" void JS_uiElemOpenStart(const char *tag);
extern "C" void JS_uiElemOpenStartKeyInt(const char *tag, int key);
extern "C" void JS_uiElemOpenStartKeyString(const char *tag, const char *key);
extern "C" void JS_uiElemOpenEnd();
extern "C" void JS_uiElemClose(const char *tag);
extern "C" int JS_uiGetToken();

// Attributes
extern "C" void JS_uiAttrInt(const char *name, int value);
extern "C" void JS_uiAttrFloat(const char *name, float value);
extern "C" void JS_uiAttrDouble(const char *name, double value);
extern "C" void JS_uiAttrString(const char *name, const char *value);
extern "C" void JS_uiAttrEmpty(const char *name);
extern "C" void JS_uiAttrClass(const char *value);

// Text
extern "C" void JS_uiText(const char *value);

// Events
extern "C" int JS_uiGetEventCount(const char *type);
extern "C" void JS_uiClearEventCounts();
extern "C" bool JS_uiIsKeyboardCaptured();

// Value
extern "C" char *JS_uiGetValue();
extern "C" void JS_uiSetValue(const char *value);

// Patch
extern "C" void JS_uiPatch(const char *id);


//
// Events
//

inline bool isModifierKeyDown() {
  for (auto mod = rl::KEY_LEFT_SHIFT; mod <= rl::KEY_RIGHT_SUPER;
       mod = rl::KeyboardKey(int(mod) + 1)) {
    if (rl::IsKeyDown(mod)) {
      return true;
    }
  }
  return false;
}

inline bool uiIsKeyboardCaptured() {
  return JS_uiIsKeyboardCaptured();
}


//
// Elements
//

struct UIElem {
  const char *tag = "div";
  bool ended = false;

  ~UIElem() {
    end();
    JS_uiElemClose(tag);
  }

  UIElem &operator()(const char *attrName, int value) {
    JS_uiAttrInt(attrName, value);
    return *this;
  }

  UIElem &operator()(const char *attrName, float value) {
    JS_uiAttrFloat(attrName, value);
    return *this;
  }

  UIElem &operator()(const char *attrName, double value) {
    JS_uiAttrDouble(attrName, value);
    return *this;
  }

  UIElem &operator()(const char *attrName, const char *value) {
    JS_uiAttrString(attrName, value);
    return *this;
  }

  UIElem &operator()(const char *attrName, char *value) {
    return operator()(attrName, (const char *)value);
  }

  UIElem &operator()(const char *attrName, bool value) {
    if (value) {
      JS_uiAttrEmpty(attrName);
    }
    return *this;
  }

  UIElem &operator()(const char *klass) {
    JS_uiAttrClass(klass);
    return *this;
  }

  UIElem &operator()(char *klass) {
    JS_uiAttrClass(klass);
    return *this;
  }

  UIElem &operator()(const char *eventName, auto &&f) {
    end();
    if (JS_uiGetEventCount(eventName) > 0) {
      callEventHandler(f);
    }
    return *this;
  }

  UIElem &operator()(const char *eventName, rl::KeyboardKey key, auto &&f) {
    if (!JS_uiIsKeyboardCaptured() && rl::IsKeyReleased(key)) {
      if (!isModifierKeyDown()) {
        callEventHandler(f);
      }
    }
    return operator()(eventName, std::forward<decltype(f)>(f));
  }

  UIElem &operator()(const char *eventName, rl::KeyboardKey mod, rl::KeyboardKey key, auto &&f) {
    if (!JS_uiIsKeyboardCaptured()
        && ((rl::IsKeyDown(mod) && rl::IsKeyReleased(key))
            || (rl::IsKeyReleased(mod) && rl::IsKeyDown(key)))) {
      callEventHandler(f);
    }
    return operator()(eventName, std::forward<decltype(f)>(f));
  }

  UIElem &operator()(auto &&f) {
    end();
    f();
    return *this;
  }

  void end() {
    if (!ended) {
      JS_uiElemOpenEnd();
      ended = true;
    }
  }

  template<typename F>
  void callEventHandler(F &&f) {
    if constexpr (std::is_invocable_v<F, const char *>) {
      Scoped value { JS_uiGetValue() };
      if constexpr (std::is_convertible_v<std::invoke_result_t<F, const char *>, const char *>) {
        JS_uiSetValue(f(value));
      } else {
        f(value);
      }
    } else {
      f();
    }
  }
};

inline UIElem ui(const char *tag) {
  JS_uiElemOpenStart(tag);
  return { .tag = tag };
}

inline UIElem ui(const char *tag, int key) {
  JS_uiElemOpenStartKeyInt(tag, key);
  return { .tag = tag };
}

inline UIElem ui(const char *tag, const char *key) {
  JS_uiElemOpenStartKeyString(tag, key);
  return { .tag = tag };
}

inline int uiGetToken() {
  return JS_uiGetToken();
}


//
// Text
//

inline void uiText(const char *format, auto &&...args) {
  JS_uiText(bprint<1024>(format, std::forward<decltype(args)>(args)...));
}

inline void uiText(const char *msg) {
  JS_uiText(msg);
}


//
// Patch
//

inline struct UIPatchState {
  void (*wrapper)() = nullptr;
  void *func = nullptr;
} uiPatchState;

template<typename F>
inline void uiPatch(const char *id, F &&f) {
#ifdef __EMSCRIPTEN__
  uiPatchState.func = (void *)&f;
  uiPatchState.wrapper = []() {
    if (uiPatchState.func) {
      auto &f = *static_cast<F *>(uiPatchState.func);
      f();
    }
  };
  JS_uiPatch(id);
  uiPatchState = {};
#else
  f();
#endif
}
