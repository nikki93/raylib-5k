#include "core/core.hh"

#include "core/game.hh"

#define CUTE_C2_IMPLEMENTATION
#include "core/c2.h"


//
// Game-defined
//

void initGame();
void updateGame(float dt);
void drawGame();


//
// Frame
//

void frame() {
  // Window size
  Vec2 windowSize { float(rl::GetScreenWidth()), float(rl::GetScreenHeight()) };
  bool windowResized [[maybe_unused]] = false;
  {
#ifdef __EMSCRIPTEN__
    Vec2 canvasSize {
      float(int(EM_ASM_DOUBLE({
        return window.devicePixelRatio
            * document.querySelector("#canvas").getBoundingClientRect().width;
      }))),
      float(int(EM_ASM_DOUBLE({
        return window.devicePixelRatio
            * document.querySelector("#canvas").getBoundingClientRect().height;
      }))),
    };
    if (windowSize != canvasSize) {
      rl::SetWindowSize(int(canvasSize.x), int(canvasSize.y));
      windowResized = true;
    }
#endif
  }

  // Pause on window unfocus on web
  {
#ifdef __EMSCRIPTEN__
    static auto firstFrame = true;
    if (!firstFrame) {
      static auto prevFocused = true;
      bool focused = EM_ASM_INT({ return document.hasFocus() ? 1 : 0; });
      if (focused != prevFocused) {
        prevFocused = focused;
        if (focused) {
          emscripten_set_main_loop_timing(EM_TIMING_RAF, 0);
        } else {
          emscripten_set_main_loop_timing(EM_TIMING_SETTIMEOUT, 100);
        }
      }
      if (!windowResized && !focused) {
        return;
      }
    }
    firstFrame = false;
#endif
  }

  // Debug display
  clearDebugDisplay();

  // UI
  uiEdit();

  // Update
  if (edit.enabled) {
    updateEdit();
  } else {
    updateGame(std::min(rl::GetFrameTime(), 0.06f));
  }

  // Draw
  rl::BeginDrawing();
  {
    rl::ClearBackground({ 0xcc, 0xe4, 0xf5, 0xff });

    gameCamera.offset = 0.5 * windowSize;
    gameCamera.zoom = float(windowSize.x) / gameCameraSize.x;
    rl::BeginMode2D(edit.enabled ? edit.camera : gameCamera);
    drawGame();
    if (edit.enabled) {
      drawEdit();
    }
    rl::EndMode2D();

    rl::DrawText(debugDisplayBuffer, 18, 18, 30, rl::WHITE);
  }
  rl::EndDrawing();

  // Flush console
  std::fflush(stdout);
}


//
// Main
//

int main() {
  //if (!loadEditSession()) {
    initGame();
  //}

#ifdef __EMSCRIPTEN__
  emscripten_set_main_loop(frame, 0, true);
#else
  while (!rl::WindowShouldClose()) {
    frame();
  }
#endif

  return 0;
}
