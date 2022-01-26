#pragma once
#include "core/core.hh"


inline void read(Props auto &val, const cJSON *jsn); // Forward declaration
inline cJSON *write(const Props auto &val);


//
// Assets
//

inline const char *getAssetPath(const char *assetName) {
  return bprint<128>("assets/%s", assetName);
}

inline const char *getAssetContents(const char *assetName) {
  if (Scoped file { std::fopen(getAssetPath(assetName), "r") }) {
    std::fseek(file, 0, SEEK_END);
    auto size = std::ftell(file);
    static char buf[6 * 1024 * 1024];
    if (long(sizeof(buf)) < size + 1) {
      cprint("asset '%s' is too large to read into memory!\n", assetName);
      return "";
    }
    std::rewind(file);
    std::fread(buf, 1, size, file);
    buf[size] = '\0';
    return buf;
  } else {
    cprint("asset '%s' could not be read: %s\n", assetName, std::strerror(errno));
    return "";
  }
}

inline void writeAssetContents(const char *assetName, const char *contents) {
  if (Scoped file { std::fopen(getAssetPath(assetName), "w") }) {
    std::fputs(contents, file);
  } else {
    cprint("asset '%s' could not be written: %s\n", assetName, std::strerror(errno));
  }
}


//
// Primitives
//

// int
inline void read(int &val, const cJSON *jsn) {
  if (cJSON_IsNumber(jsn)) {
    val = jsn->valueint;
  }
}
inline cJSON *write(const int &val) {
  return cJSON_CreateNumber(val);
}

// float
inline void read(float &val, const cJSON *jsn) {
  if (cJSON_IsNumber(jsn)) {
    val = float(jsn->valuedouble);
  }
}
inline cJSON *write(const float &val) {
  return cJSON_CreateNumber(val);
}

// double
inline void read(double &val, const cJSON *jsn) {
  if (cJSON_IsNumber(jsn)) {
    val = jsn->valuedouble;
  }
}
inline cJSON *write(const double &val) {
  return cJSON_CreateNumber(val);
}

// bool
inline void read(bool &val, const cJSON *jsn) {
  if (cJSON_IsBool(jsn)) {
    val = cJSON_IsTrue(jsn);
  }
}
inline cJSON *write(const bool &val) {
  return cJSON_CreateBool(val);
}

// char array
template<int N>
void read(char (&val)[N], const cJSON *jsn) {
  if (cJSON_IsString(jsn)) {
    copy(val, cJSON_GetStringValue(jsn));
  }
}
inline cJSON *write(const char *val) {
  return cJSON_CreateString(val);
}

// gx::String
inline void read(gx::String &val, const cJSON *jsn) {
  if (cJSON_IsString(jsn)) {
    val = cJSON_GetStringValue(jsn);
  }
}
inline cJSON *write(const gx::String &val) {
  return cJSON_CreateString(val);
}

// Entity
inline void read(Entity &val, const cJSON *jsn) {
  if (cJSON_IsNumber(jsn)) {
    val = Entity(jsn->valueint);
  }
}
inline cJSON *write(const Entity &val) {
  return cJSON_CreateNumber(double(uint32_t(val)));
}


//
// Graphics
//

// PERF: Can use direct access to nodes -- array get scans

// Vec2
inline void read(Vec2 &val, const cJSON *jsn) {
  if (cJSON_IsArray(jsn) && cJSON_GetArraySize(jsn) == 2) {
    read(val.x, cJSON_GetArrayItem(jsn, 0));
    read(val.y, cJSON_GetArrayItem(jsn, 1));
  }
}
inline cJSON *write(const Vec2 &val) {
  auto result = cJSON_CreateArray();
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.x));
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.y));
  return result;
}

// rl::Rectangle
inline void read(rl::Rectangle &val, const cJSON *jsn) {
  if (cJSON_IsArray(jsn) && cJSON_GetArraySize(jsn) == 4) {
    read(val.x, cJSON_GetArrayItem(jsn, 0));
    read(val.y, cJSON_GetArrayItem(jsn, 1));
    read(val.width, cJSON_GetArrayItem(jsn, 2));
    read(val.height, cJSON_GetArrayItem(jsn, 3));
  }
}
inline cJSON *write(const rl::Rectangle &val) {
  auto result = cJSON_CreateArray();
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.x));
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.y));
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.width));
  cJSON_AddItemToArray(result, cJSON_CreateNumber(val.height));
  return result;
}

// rl::Camera2D
struct ReadWriteCamera {
  Prop(Vec2, offset);
  Prop(Vec2, target);
  Prop(float, rotation);
  Prop(float, zoom);
};
inline void read(rl::Camera2D &val, const cJSON *jsn) {
  ReadWriteCamera rw {
    .offset = val.offset, .target = val.target, .rotation = val.rotation, .zoom = val.zoom
  };
  read(rw, jsn);
  val = { .offset = rw.offset, .target = rw.target, .rotation = rw.rotation, .zoom = rw.zoom };
}
inline cJSON *write(const rl::Camera2D &val) {
  return write(ReadWriteCamera {
      .offset = val.offset, .target = val.target, .rotation = val.rotation, .zoom = val.zoom });
}


//
// Containers
//

// T[N]
template<typename T, unsigned N>
void read(T (&val)[N], const cJSON *jsn) {
  if (cJSON_IsArray(jsn)) {
    auto i = 0;
    for (auto elemJsn = jsn->child; elemJsn; elemJsn = elemJsn->next) {
      if (i >= int(N)) {
        break;
      }
      read(val[i++], elemJsn);
    }
  }
}
template<typename T, unsigned N>
cJSON *write(const T (&val)[N]) {
  auto result = cJSON_CreateArray();
  for (auto &elem : val) {
    cJSON_AddItemToArray(result, write(elem));
  }
  return result;
}

// gx::Array<T, N>
template<typename T, int N>
void read(gx::Array<T, N> &val, const cJSON *jsn) {
  if (cJSON_IsArray(jsn)) {
    auto i = 0;
    for (auto elemJsn = jsn->child; elemJsn; elemJsn = elemJsn->next) {
      if (i >= int(N)) {
        break;
      }
      read(val[i++], elemJsn);
    }
  }
}
template<typename T, int N>
cJSON *write(const gx::Array<T, N> &val) {
  auto result = cJSON_CreateArray();
  for (auto &elem : val) {
    cJSON_AddItemToArray(result, write(elem));
  }
  return result;
}

// gx::Slice<T>
template<typename T>
inline void read(gx::Slice<T> &val, const cJSON *jsn) {
  if (cJSON_IsArray(jsn)) {
    auto size = cJSON_GetArraySize(jsn);
    val = {};
    val.data = (T *)std::malloc(sizeof(T) * size);
    val.size = size;
    val.capacity = size;
    auto i = 0;
    for (auto elemJsn = jsn->child; elemJsn; elemJsn = elemJsn->next) {
      new (&val.data[i]) T();
      read(val[i], elemJsn);
      ++i;
    }
  }
}
template<typename T>
inline cJSON *write(const gx::Slice<T> &val) {
  auto result = cJSON_CreateArray();
  for (auto &elem : val) {
    cJSON_AddItemToArray(result, write(elem));
  }
  return result;
}

// Scoped<T>
template<typename T>
inline void read(Scoped<T> &val, const cJSON *jsn) {
  read(val.val, jsn);
}
template<typename T>
inline cJSON *write(const Scoped<T> &val) {
  return write(val.val);
}

// cJSON *
inline void read(cJSON *&val, const cJSON *jsn) {
  val = cJSON_Duplicate(jsn, true);
}
inline cJSON *write(cJSON *const &jsn) {
  return cJSON_Duplicate(jsn, true);
}


//
// Props
//

void read(Props auto &val, const cJSON *jsn) {
  if (cJSON_IsObject(jsn)) {
    for (auto elemJsn = jsn->child; elemJsn; elemJsn = elemJsn->next) {
      const auto key = elemJsn->string;
      const auto keyHash = hash(key);
      forEachProp(val, [&](auto propTag, auto &propVal) {
        constexpr auto propNameHash = propTag.attribs.nameHash;
        constexpr auto propName = propTag.attribs.name;
        if (keyHash == propNameHash && key == propName) {
          read(propVal, elemJsn);
        }
      });
    }
  }
}
inline cJSON *write(const Props auto &val) {
  auto result = cJSON_CreateObject();
  forEachProp(val, [&](auto propTag, auto &propVal) {
    cJSON_AddItemToObjectCS(result, propTag.attribs.name.data(), write(propVal));
  });
  return result;
}


//
// `cJSON *` -> `const char *`
//

inline const char *stringify(cJSON *jsn, bool formatted = false) {
  static char buf[6 * 1024 * 1024];
  if (!cJSON_PrintPreallocated(jsn, buf, sizeof(buf), formatted)) {
    cprint("json is too large to stringify!");
    return "";
  }
  return buf;
}


//
// Scene
//

// Blueprint
Entity readBlueprint(const cJSON *jsn);
Entity readBlueprint(const char *assetName);
cJSON *writeBlueprint(Entity ent, bool writeId = false);

// Top-level
void readScene(const cJSON *jsn);
void readScene(const char *assetName);
cJSON *writeScene();
