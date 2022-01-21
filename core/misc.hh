#pragma once
#include "core/core.hh"


//
// Strings
//

auto printfArg(auto arg) {
  return arg;
}

inline const char *printfArg(const gx::String &s) {
  return s;
}

template<int N>
void copy(char (&dest)[N], const char *src) {
  if constexpr (N > 0) {
    dest[0] = '\0';
    std::strncat(dest, src, N - 1);
  }
}

template<int N>
const char *bprint(char (&buf)[N], const char *format, auto &&...args) {
  std::snprintf(buf, N, format, printfArg(args)...);
  return buf;
}

template<int N>
const char *bprint(char (&buf)[N], const char *str) {
  copy(buf, str);
  return buf;
}

template<int N>
const char *bprint(const char *format, auto &&...args) {
  static char buf[N];
  return bprint(buf, format, printfArg(args)...);
}

template<int N>
const char *nullTerminate(std::string_view str) {
  static char buf[N];
  buf[0] = '\0';
  std::strncat(buf, str.data(), std::min(N - 1, int(str.size())));
  buf[str.size()] = '\0';
  return buf;
}

inline bool isEmpty(const char *str) {
  return str[0] == '\0';
}

template<int N>
inline void clear(char (&str)[N]) {
  str[0] = '\0';
}

template<int N>
inline bool endsWith(const char *str, const char (&suffix)[N]) {
  constexpr auto suffixLen = N - 1;
  if (auto strLen = std::strlen(str); strLen >= suffixLen) {
    return !std::strcmp(&str[strLen - suffixLen], suffix);
  } else {
    return false;
  }
}


//
// Hashing
//

using HashedString = entt::hashed_string;
using namespace entt::literals;

constexpr uint32_t hash(const char *str) {
  return entt::hashed_string(str).value();
}

constexpr uint32_t hash(std::string_view str) {
  return entt::hashed_string::value(str.data(), str.size());
}


//
// Release
//

// char *
inline void release(char *&val) {
  std::free(val);
  val = nullptr;
}

// FILE *
inline void release(FILE *&val) {
  if (val) {
    std::fclose(val);
    val = nullptr;
  }
}

// cJSON *
inline void release(cJSON *&val) {
  cJSON_Delete(val);
  val = nullptr;
}

// Scoped
template<typename T>
struct Scoped {
  T val {};

  Scoped() = default;

  Scoped(const Scoped &) = delete;
  Scoped &operator=(const Scoped &) = delete;

  explicit Scoped(T val_)
      : val(std::move(val_)) {
  }

  ~Scoped() {
    release(val);
  }

  operator T &() { // NOLINT(google-explicit-constructor)
    return val;
  }
};


//
// Console
//

inline void cprint(const char *format, auto &&...args) {
  std::printf(format, printfArg(args)...);
  std::puts("");
}

inline void cprint(const char *msg) {
  std::puts(msg);
}


//
// Debug display
//

inline char debugDisplayBuffer[1024];
inline char *debugDisplayCursor = debugDisplayBuffer;

inline void clearDebugDisplay() {
  clear(debugDisplayBuffer);
  debugDisplayCursor = debugDisplayBuffer;
}

inline void dprint(const char *format, auto &&...args) {
  auto remaining = int(sizeof(debugDisplayBuffer) - (debugDisplayCursor - debugDisplayBuffer));
  if (remaining > 0) {
    auto expected = std::snprintf(debugDisplayCursor, remaining, format, printfArg(args)...);
    debugDisplayCursor += std::min(expected, remaining);
    if (remaining - expected > 2) {
      *debugDisplayCursor++ = '\n';
      *debugDisplayCursor = '\0';
    }
  }
}

inline void dprint(const char *msg) {
  dprint("%s", msg);
}


//
// Leak check
//

#if defined(__has_feature) && __has_feature(address_sanitizer)
constexpr auto leakCheckAvailable = true;
namespace __lsan {
void DoRecoverableLeakCheckVoid();
void DisableInThisThread();
void EnableInThisThread();
}
#else
constexpr auto leakCheckAvailable = false;
namespace __lsan {
inline void DoRecoverableLeakCheckVoid() {
}
inline void DisableInThisThread() {
}
inline void EnableInThisThread() {
}
}
#endif
