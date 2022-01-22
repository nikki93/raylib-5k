#include "core/core.hh"

#include "core/game.hh"
UseComponentTypes();


//
// History
//

struct EditComponentSnapshot {
  void *data;
  void (*free)(void *data);
  void (*add)(void *data, Entity ent);
  cJSON *(*write)(void *data);
};
struct EditEntitySnapshot {
  Prop((Seq<EditComponentSnapshot, 8>), componentSnapshots);
  Prop(Entity, ent);
};
struct EditSceneSnapshot {
  Prop(bool, valid) = false;
  Prop(double, time) = 0;
  Prop(char[64], description);
  Prop(char[64], inspectedComponentTitle) = "";
  Prop(Seq<EditEntitySnapshot>, entitySnapshots);
  Prop(Seq<Entity>, selection);
};
static constexpr auto editMaxSnapshots = 50;
static struct EditHistory {
  Prop(EditSceneSnapshot[editMaxSnapshots], snapshots);
  Prop(int, lastIndex) = 0;
  Prop(Seq<Entity>, stopSelection);
} editHistory;


// Component snapshot

template<typename T>
void init(EditComponentSnapshot &componentSnapshot) {
  componentSnapshot.data = new T {};
  componentSnapshot.free = [](void *data) {
    delete ((T *)data);
  };
  componentSnapshot.add = [](void *data, Entity ent) {
    auto copy = *((T *)data);
    add<T>(ent, std::move(copy));
  };
  componentSnapshot.write = [](void *data) {
    auto &comp = *((T *)data);
    auto jsn = write(comp);
    cJSON_AddItemToObjectCS(jsn, "_type", cJSON_CreateString(nullTerminate<64>(getTypeName<T>())));
    return jsn;
  };
}

void read(EditComponentSnapshot &componentSnapshot, const cJSON *jsn) {
  const auto typeName = cJSON_GetStringValue(cJSON_GetObjectItemCaseSensitive(jsn, "_type"));
  const auto typeNameHash = hash(typeName);
  forEachComponentType([&]<typename T>() {
    constexpr auto TName = getTypeName<T>();
    constexpr auto TNameHash = hash(TName);
    if (typeNameHash == TNameHash && typeName == TName) {
      init<T>(componentSnapshot);
      T &comp = *((T *)componentSnapshot.data);
      read(comp, jsn);
    }
  });
}

cJSON *write(const EditComponentSnapshot &componentSnapshot) {
  return componentSnapshot.write(componentSnapshot.data);
}


// Selection

static Seq<Entity> saveEditSelection() {
  Seq<Entity> result;
  each([&](Entity ent, EditSelect &sel) {
    append(result, ent);
  });
  return result;
}

static void loadEditSelection(const Seq<Entity> &selection) {
  clear<EditSelect>();
  for (auto ent : selection) {
    if (exists(ent)) {
      add<EditSelect>(ent);
    }
  }
}


// Reset

static void reset(EditSceneSnapshot &snapshot) {
  for (auto &entitySnapshot : snapshot.entitySnapshots) {
    for (auto &componentSnapshot : entitySnapshot.componentSnapshots) {
      componentSnapshot.free(componentSnapshot.data);
    }
  }
  snapshot = {};
}

void resetEditHistory() {
  for (auto &snapshot : editHistory.snapshots) {
    reset(snapshot);
  }
  editHistory = {};
}

static struct EditResetHistoryOnExit {
  ~EditResetHistoryOnExit() {
    resetEditHistory();
  }
} editResetHistoryOnExit;


// Save / load

void saveEditSnapshot(const char *desc, bool saveInspectedComponentTitle) {
  if (!edit.enabled) {
    return;
  }

  validateGameEdit();

  clear(edit.notification);

  editHistory.lastIndex = (editHistory.lastIndex + 1) % editMaxSnapshots;

  auto &snapshot = editHistory.snapshots[editHistory.lastIndex];
  reset(snapshot);
  snapshot.valid = true;
  snapshot.time = rl::GetTime();
  copy(snapshot.description, desc);
  if (saveInspectedComponentTitle) {
    copy(snapshot.inspectedComponentTitle, edit.inspectedComponentTitle);
  }

  Seq<Entity> entities;
  each([&](Entity ent) {
    append(entities, ent);
  });
  std::reverse(entities.begin(), entities.end());
  for (auto ent : entities) {
    if (!has<EditDelete>(ent)) {
      auto &entitySnapshot = append(snapshot.entitySnapshots);
      entitySnapshot.ent = ent;
      forEachComponentType([&]<typename T>() {
        if (has<T>(ent)) {
          auto &componentSnapshot = append(entitySnapshot.componentSnapshots);
          init<T>(componentSnapshot);
          T &destComp = *((T *)componentSnapshot.data);
          auto &srcComp = get<T>(ent);
          forEachProp(destComp, [&]<typename DestTag, typename Prop>(DestTag, Prop &destProp) {
            forEachProp(srcComp, [&]<typename SrcTag>(SrcTag, auto &srcProp) {
              if constexpr (std::is_same_v<DestTag, SrcTag>) {
                if constexpr (requires { destProp = srcProp; }) {
                  destProp = srcProp;
                } else {
                  std::memcpy(&destProp, &srcProp, sizeof(Prop)); // TODO: Ensure `char[N]`
                }
              }
            });
          });
        }
      });
    }
  }

  snapshot.selection = saveEditSelection();
}

static void loadEditSnapshot() {
  if (auto &snapshot = editHistory.snapshots[editHistory.lastIndex]; snapshot.valid) {
    each(destroyEntity);
    for (auto &entitySnapshot : snapshot.entitySnapshots) {
      auto ent = createEntity(entitySnapshot.ent);
      for (auto &componentSnapshot : entitySnapshot.componentSnapshots) {
        componentSnapshot.add(componentSnapshot.data, ent);
      }
    }
  }
}


// Undo / redo

bool canUndoEdit() {
  auto &curr = editHistory.snapshots[editHistory.lastIndex];
  auto prevI = editHistory.lastIndex == 0 ? editMaxSnapshots - 1 : editHistory.lastIndex - 1;
  auto &prev = editHistory.snapshots[prevI];
  return prev.valid && (!curr.valid || prev.time <= curr.time);
}

void undoEdit() {
  auto &curr = editHistory.snapshots[editHistory.lastIndex];
  auto prevI = editHistory.lastIndex == 0 ? editMaxSnapshots - 1 : editHistory.lastIndex - 1;
  auto &prev = editHistory.snapshots[prevI];
  if (prev.valid && (!curr.valid || prev.time <= curr.time)) {
    editHistory.lastIndex = prevI;
    loadEditSnapshot();
    if (curr.valid) {
      loadEditSelection(curr.selection);
      if (!isEmpty(curr.inspectedComponentTitle)) {
        copy(edit.inspectedComponentTitle, curr.inspectedComponentTitle);
      }
      notifyEdit("undid %s", curr.description);
    }
  }
}

bool canRedoEdit() {
  auto &curr = editHistory.snapshots[editHistory.lastIndex];
  auto nextI = editHistory.lastIndex == editMaxSnapshots - 1 ? 0 : editHistory.lastIndex + 1;
  auto &next = editHistory.snapshots[nextI];
  return (next.valid && (!curr.valid || next.time >= curr.time));
}

void redoEdit() {
  auto &curr = editHistory.snapshots[editHistory.lastIndex];
  auto nextI = editHistory.lastIndex == editMaxSnapshots - 1 ? 0 : editHistory.lastIndex + 1;
  auto &next = editHistory.snapshots[nextI];
  if (next.valid && (!curr.valid || next.time >= curr.time)) {
    editHistory.lastIndex = nextI;
    loadEditSnapshot();
    loadEditSelection(next.selection);
    if (!isEmpty(next.inspectedComponentTitle)) {
      copy(edit.inspectedComponentTitle, next.inspectedComponentTitle);
    }
    notifyEdit("redid %s", next.description);
  }
}


//
// Session
//

struct EditSession {
  Prop(Scoped<cJSON *>, scene);
  Prop(EditState, edit);
  Prop(EditHistory, history);
  Prop(Seq<Entity>, selection);
};

JS_DEFINE(char *, JS_getStorage, (const char *key), {
  let found = localStorage.getItem(UTF8ToString(key));
  if (!found) {
    found = "";
  }
  return allocate(intArrayFromString(found), ALLOC_NORMAL);
});

JS_DEFINE(void, JS_setStorage, (const char *key, const char *value),
    { localStorage.setItem(UTF8ToString(key), UTF8ToString(value)); });

bool loadEditSession() {
  Scoped sessionStr { JS_getStorage("editSession") };
  JS_setStorage("editSession", "");
  if (!sessionStr || isEmpty(sessionStr)) {
    return false;
  }
  Scoped sessionJsn { cJSON_Parse(sessionStr) };

  EditSession session;
  read(session, sessionJsn);

  each(destroyEntity);
  readScene(session.scene);

  edit = session.edit;

  resetEditHistory();
  editHistory = std::move(session.history);
  auto maxTime = -INFINITY;
  for (auto &snapshot : editHistory.snapshots) {
    if (snapshot.valid && snapshot.time > maxTime) {
      maxTime = float(snapshot.time);
    }
  }
  for (auto &snapshot : editHistory.snapshots) {
    snapshot.time -= maxTime + 1;
  }

  loadEditSelection(session.selection);

  notifyEdit("restored session");
  return true;
}

JS_EXPORT void JS_saveEditSession() {
  EditSession session {
    .scene = Scoped { writeScene() },
    .edit = edit,
    .history = editHistory,
    .selection = saveEditSelection(),
  };
  Scoped sessionJsn { write(session) };
  JS_setStorage("editSession", stringify(sessionJsn));

  notifyEdit("saved session");
}


//
// Play / stop
//

void playEdit() {
  if (edit.enabled) {
    editHistory.stopSelection = saveEditSelection();
    edit.enabled = false;
    clear(edit.notification);
  }
}

void stopEdit() {
  if (!edit.enabled) {
    edit.enabled = true;
    loadEditSnapshot();
    loadEditSelection(editHistory.stopSelection);
    clear(edit.notification);
  }
}


//
// Scene
//

void openSceneEdit(const char *sceneName) {
  each(destroyEntity);
  resetEditHistory();
  readScene(sceneName);
  saveEditSnapshot("load scene");
  edit.camera.target = { 0, 0 };
  copy(edit.sceneName, sceneName);
}

void saveSceneEdit(const char *sceneName) {
  Scoped jsn { writeScene() };
  writeAssetContents(sceneName, stringify(jsn, true));
  copy(edit.sceneName, sceneName);
}
