#pragma once
#include "core/core.hh"


//
// Math aliases and operators
//

using Vec2 = rl::Vector2;

inline bool operator==(Vec2 u, Vec2 v) {
  return u.x == v.x && u.y == v.y;
}
inline bool operator!=(Vec2 u, Vec2 v) {
  return !(u == v);
}

inline Vec2 operator-(Vec2 v) {
  return Vec2 { -v.x, -v.y };
}
inline Vec2 operator+(Vec2 u, Vec2 v) {
  return rl::Vector2Add(u, v);
}
inline Vec2 &operator+=(Vec2 &u, Vec2 v) {
  u = u + v;
  return u;
}
inline Vec2 operator-(Vec2 u, Vec2 v) {
  return rl::Vector2Subtract(u, v);
}
inline Vec2 &operator-=(Vec2 &u, Vec2 v) {
  u = u - v;
  return u;
}
inline Vec2 operator*(Vec2 u, Vec2 v) {
  return rl::Vector2Multiply(u, v);
}
inline Vec2 &operator*=(Vec2 &u, Vec2 v) {
  u = u * v;
  return u;
}
inline Vec2 operator/(Vec2 u, Vec2 v) {
  return rl::Vector2Divide(u, v);
}
inline Vec2 &operator/=(Vec2 &u, Vec2 v) {
  u = u / v;
  return u;
}

inline Vec2 operator*(float f, Vec2 v) {
  return rl::Vector2Scale(v, f);
}
inline Vec2 operator*(Vec2 v, float f) {
  return rl::Vector2Scale(v, f);
}
inline Vec2 &operator*=(Vec2 &v, float f) {
  v = v * f;
  return v;
}
inline Vec2 operator/(Vec2 v, float f) {
  return v * (1 / f);
}
inline Vec2 &operator/=(Vec2 &v, float f) {
  v = v / f;
  return v;
}


namespace rl {


//
// RLGL function aliases
//

inline void MatrixMode(int mode) {
  rlMatrixMode(mode);
}
inline void PushMatrix() {
  rlPushMatrix();
}
inline void PopMatrix() {
  rlPopMatrix();
}
inline void LoadIdentity() {
  rlLoadIdentity();
}
inline void Translatef(float x, float y, float z) {
  rlTranslatef(x, y, z);
}
inline void Rotatef(float angleDeg, float x, float y, float z) {
  rlRotatef(angleDeg, x, y, z);
}
inline void Scalef(float x, float y, float z) {
  rlScalef(x, y, z);
}
inline void MultMatrixf(float *matf) {
  rlMultMatrixf(matf);
}
inline void Frustum(
    double left, double right, double bottom, double top, double znear, double zfar) {
  rlFrustum(left, right, bottom, top, znear, zfar);
}
inline void Ortho(double left, double right, double bottom, double top, double znear, double zfar) {
  rlOrtho(left, right, bottom, top, znear, zfar);
}
inline void Viewport(int x, int y, int width, int height) {
  rlViewport(x, y, width, height);
}
inline Matrix GetMatrixModelview() {
  return rlGetMatrixModelview();
}


//
// Init / deinit on program startup / exit -- allows creating resources at global scope
//

inline struct Init {
  Init() {
    SetTraceLogLevel(LOG_WARNING);
#ifndef __EMSCRIPTEN__
    SetConfigFlags(FLAG_VSYNC_HINT);
#endif
#ifdef __EMSCRIPTEN__
    Vec2 windowSize {
      float(int(EM_ASM_DOUBLE({
        return window.devicePixelRatio
            * document.querySelector("#canvas").getBoundingClientRect().width;
      }))),
      float(int(EM_ASM_DOUBLE({
        return window.devicePixelRatio
            * document.querySelector("#canvas").getBoundingClientRect().height;
      }))),
    };
#else
    auto windowSize = 1.5 * Vec2 { 960, 540 };
#endif
    __lsan::DisableInThisThread(); // Some Emscripten initialization things leak memory...
    InitWindow(int(windowSize.x), int(windowSize.y), "");
    __lsan::EnableInThisThread();
#ifdef __EMSCRIPTEN__
    EM_ASM({ initElectronFS(); });
#endif
  }
  ~Init() {
    CloseWindow();
  }
} init;


}
