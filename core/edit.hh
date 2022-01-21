#pragma once
#include "core/core.hh"


//
// State
//

constexpr Vec2 editInitialCameraPos { 0.5 * 960, 0.5 * 540 };

inline struct EditState {
  Prop(bool, enabled) = []() {
#ifdef __EMSCRIPTEN__
    return !bool(EM_ASM_INT({ return window.isPlayMode ? 1 : 0; }));
#else
    return false;
#endif
  }();
  Prop(char[16], mode) = "select";
  Prop(char[64], sceneName) = "";
  Prop(char[64], inspectedComponentTitle) = "";

  Prop(rl::Camera2D, camera) {
    .target = editInitialCameraPos,
  };
  Prop(float, zoomLevel) = 0;

  float lineThickness = 1;

  Vec2 mouseScreenPos { 0, 0 };
  Vec2 mouseWorldPos { 0, 0 };
  Vec2 mouseWorldDelta { 0, 0 };

  char notification[96] = "";
  double lastNotificationTime = 0;
} edit;


//
// Components
//

struct EditSelect {};

struct EditDelete {};

struct EditBox {
  rl::Rectangle rect;
};

struct EditMove {
  Vec2 delta { 0, 0 };
};


//
// Interface
//

// Mode

inline bool isEditMode(const char *mode) {
  return !std::strcmp(edit.mode, mode);
}

inline void setEditMode(const char *newMode) {
  copy(edit.mode, newMode);
}


// Boxes

inline void mergeEditBox(Entity ent, rl::Rectangle rect) {
  if (has<EditBox>(ent)) {
    auto &box = get<EditBox>(ent);
    Vec2 min {
      std::min(box.rect.x, rect.x),
      std::min(box.rect.y, rect.y),
    };
    Vec2 max {
      std::max(box.rect.x + box.rect.width, rect.x + rect.width),
      std::max(box.rect.y + box.rect.height, rect.y + rect.height),
    };
    box.rect.x = min.x;
    box.rect.y = min.y;
    box.rect.width = max.x - box.rect.x;
    box.rect.height = max.y - box.rect.y;
  } else {
    add<EditBox>(ent, { rect });
  }
}


// Moves

void applyEditMoves();


// Camera

void setEditZoomLevel(float newZoomLevel);


// Session

void resetEditHistory();
void saveEditSnapshot(const char *desc, bool saveInspectedComponentTitle = false);
bool canUndoEdit();
void undoEdit();
bool canRedoEdit();
void redoEdit();

bool loadEditSession();

void playEdit();
void stopEdit();

void openSceneEdit(const char *sceneName);
void saveSceneEdit(const char *sceneName = nullptr);


// UI

void notifyEdit(const char *format, auto &&...args) {
  bprint(edit.notification, format, std::forward<decltype(args)>(args)...);
  edit.lastNotificationTime = rl::GetTime();
}

struct EditInspectContext {
  Entity ent = nullEntity;
  const char *componentTitle = "";
  bool changed = false;
  char changeDescription[96] = "";
  PropAttribs attribs;
};


//
// Top-level
//

void updateEdit();
void drawEdit();
void uiEdit();


//
// Game-defined
//

void applyGameEditMoves();
void mergeGameEditBoxes();
void inputGameEdit();
void validateGameEdit();
void drawGameEdit();
