#include "core/core.hh"

#include "core/game.hh"


//
// Update
//

void applyEditMoves() {
  applyGameEditMoves();
  clear<EditMove>();
}

void setEditZoomLevel(float newZoomLevel) {
  edit.zoomLevel = newZoomLevel;
  if (edit.zoomLevel == 0) {
    notifyEdit("zoomed 1x");
  } else if (edit.zoomLevel > 0) {
    notifyEdit("zoomed in %.3gx", std::pow(1.5, edit.zoomLevel));
  } else {
    notifyEdit("zoomed out %.3gx", 1 / std::pow(1.5, edit.zoomLevel));
  }
}

void updateEdit() {
  // Mouse coordinates
  auto prevMouseScreenPos = edit.mouseScreenPos;
  edit.mouseScreenPos = rl::GetMousePosition();
  auto prevMouseWorldPos = rl::GetScreenToWorld2D(prevMouseScreenPos, edit.camera);
  edit.mouseWorldPos = rl::GetScreenToWorld2D(edit.mouseScreenPos, edit.camera);
  if (rl::IsMouseButtonPressed(0)) {
    prevMouseScreenPos = edit.mouseScreenPos;
    prevMouseWorldPos = edit.mouseWorldPos;
  }
  edit.mouseWorldDelta = edit.mouseWorldPos - prevMouseWorldPos;

  // Camera pan
  {
    static char prevMode[sizeof(edit.mode)] = "select";
    if (rl::IsMouseButtonPressed(2)) {
      copy(prevMode, edit.mode);
      setEditMode("camera pan");
    }
    if (rl::IsMouseButtonReleased(2)) {
      copy(edit.mode, prevMode);
    }
    if (isEditMode("camera pan")) {
      if (rl::IsMouseButtonDown(0) || rl::IsMouseButtonDown(2)) {
        edit.camera.target -= edit.mouseWorldDelta;
      }
    }
  }

  // Camera zoom and offset
  {
    static float wheelMove = 0;
    wheelMove += rl::GetMouseWheelMove();
    if (std::abs(wheelMove) > 0.1) {
      if (wheelMove < 0) {
        setEditZoomLevel(edit.zoomLevel + 1);
      } else {
        setEditZoomLevel(edit.zoomLevel - 1);
      }
      wheelMove = 0;
    }
    auto baseZoom = float(rl::GetScreenWidth()) / gameCameraSize.x;
    edit.camera.zoom = float(baseZoom * std::pow(1.5, edit.zoomLevel));
    edit.camera.offset = 0.5 * Vec2 { float(rl::GetScreenWidth()), float(rl::GetScreenHeight()) };
  }

  // Line thickness
  {
#ifdef __EMSCRIPTEN__
    static auto devicePixelRatio = []() {
      return float(EM_ASM_DOUBLE({ return window.devicePixelRatio; }));
    }();
#else
    static auto devicePixelRatio = rl::GetWindowScaleDPI().x;
#endif
    edit.lineThickness = devicePixelRatio / edit.camera.zoom;
  }

  // Deselect all
  if (!uiIsKeyboardCaptured()
      && (rl::IsKeyDown(rl::KEY_LEFT_CONTROL) && rl::IsKeyReleased(rl::KEY_D))) {
    clear<EditSelect>();
    setEditMode("select");
  }

  // Select
  if (isEditMode("select")) {
    if (rl::IsMouseButtonReleased(0)) {
      // Collect hits in ascending area order
      constexpr auto maxNumHits = 32;
      struct Hit {
        Entity ent;
        float order;
      } hits[maxNumHits];
      auto numHits = 0;
      each([&](Entity ent, EditBox &box) {
        if (numHits < maxNumHits && rl::CheckCollisionPointRec(edit.mouseWorldPos, box.rect)) {
          hits[numHits++] = { .ent = ent, .order = box.rect.width * box.rect.height };
        }
      });
      std::sort(hits, hits + numHits, [](const Hit &a, const Hit &b) {
        return a.order < b.order;
      });

      // Pick after current selection, or first if none
      Entity pick = nullEntity;
      auto pickNext = true;
      for (auto i = 0; i < numHits; ++i) {
        auto ent = hits[i].ent;
        if (has<EditSelect>(ent)) {
          pickNext = true;
        } else if (pickNext) {
          pick = ent;
          pickNext = false;
        }
      }
      if (!rl::IsKeyDown(rl::KEY_LEFT_CONTROL)) {
        clear<EditSelect>();
      }
      if (pick != nullEntity) {
        add<EditSelect>(pick);
      }
    }
  }

  // Move
  if (isEditMode("move")) {
    static auto needSnapshot = false;
    if (rl::IsMouseButtonDown(0) && edit.mouseWorldDelta != Vec2 { 0, 0 }) {
      each([&](Entity ent, EditSelect &sel, EditBox &box) {
        add<EditMove>(ent, { .delta { edit.mouseWorldDelta } });
      });
      applyEditMoves();
      needSnapshot = true;
    }
    if (rl::IsMouseButtonReleased(0) && needSnapshot) {
      saveEditSnapshot("move");
      needSnapshot = false;
    }
  }

  // Game input
  inputGameEdit();

  // Game boxes
  clear<EditBox>();
  mergeGameEditBoxes();

  // Leak check
  if constexpr (leakCheckAvailable) {
    if (!uiIsKeyboardCaptured() && rl::IsKeyReleased(rl::KEY_L)) {
      cprint("running leak check...");
      __lsan::DoRecoverableLeakCheckVoid();
      cprint("leak check done!");
    }
  }
}


//
// Draw
//

void drawEdit() {
  // Boxes
  {
    const auto drawRect = [&](rl::Rectangle rect, rl::Color color) {
      Vec2 v0 { rect.x, rect.y };
      Vec2 v1 { v0.x + rect.width, v0.y };
      Vec2 v2 { rect.x + rect.width, rect.y + rect.height };
      Vec2 v3 { v0.x, v0.y + rect.height };
      rl::DrawLineEx(v0, v1, edit.lineThickness, color);
      rl::DrawLineEx(v1, v2, edit.lineThickness, color);
      rl::DrawLineEx(v2, v3, edit.lineThickness, color);
      rl::DrawLineEx(v3, v0, edit.lineThickness, color);
    };

    // Gray for unselected entities
    each([&](Entity ent, EditBox &box) {
      if (!has<EditSelect>(ent)) {
        drawRect(box.rect, { 0x90, 0x90, 0x90, 0x80 });
      }
    });

    // Doubled green for selected entities
    constexpr rl::Color selectedColor { 0, 0x80, 0x40, 0xa0 };
    auto camTopLeft = rl::GetScreenToWorld2D({ 0, 0 }, edit.camera);
    Vec2 screenSize { float(rl::GetScreenWidth()), float(rl::GetScreenHeight()) };
    auto camBottomRight = rl::GetScreenToWorld2D(screenSize, edit.camera);
    auto border = 2 * edit.lineThickness;
    rl::Rectangle camRect {
      camTopLeft.x + border,
      camTopLeft.y + border,
      camBottomRight.x - camTopLeft.x - 2 * border,
      camBottomRight.y - camTopLeft.y - 2 * border,
    };
    each([&](Entity ent, EditSelect &sel, EditBox &box) {
      auto rect = rl::GetCollisionRec(camRect, box.rect);
      drawRect(rect, selectedColor);
      rl::Rectangle biggerRect {
        rect.x - border,
        rect.y - border,
        rect.width + 2 * border,
        rect.height + 2 * border,
      };
      drawRect(biggerRect, selectedColor);
    });
  }

  // Game hook
  drawGameEdit();
}
