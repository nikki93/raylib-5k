#include "core/core.hh"

#include "core/game.hh"
UseComponentTypes();


//
// Assets
//

JS_DEFINE(char *, JS_getAssetBlobUrl, (const char *assetName), {
  const assetNameStr = UTF8ToString(assetName);
  const bytes = FS.readFile("assets/" + assetNameStr);
  const options = {};
  options.type = "image/png";
  const blob = new Blob([bytes], options);
  const url = URL.createObjectURL(blob);
  if (!url) {
    url = "(failed)";
  }
  return allocate(intArrayFromString(url), ALLOC_NORMAL);
});

JS_DEFINE(void, JS_revokeBlobUrl, (const char *url), { URL.revokeObjectURL(UTF8ToString(url)); });


struct AssetBrowserEntry {
  char name[64] = "";
  uint32_t nameHash;
  char blobUrl[96] = "";
};

static auto getAssetBrowserEntries() {
  Seq<AssetBrowserEntry, 128> result;
  auto nDirFiles = 0;
  auto dirFiles = rl::GetDirectoryFiles("assets", &nDirFiles);
  for (auto i = 0; i < nDirFiles; ++i) {
    if (!(!std::strcmp(dirFiles[i], ".") || !std::strcmp(dirFiles[i], ".."))) {
      auto &entry = append(result);
      copy(entry.name, dirFiles[i]);
      entry.nameHash = hash(entry.name);
    }
  }
  rl::ClearDirectoryFiles();
  return result;
};

static auto assetBrowserEntries = getAssetBrowserEntries();

static void reloadAssetBrowserEntries() {
  for (auto &entry : assetBrowserEntries) {
    if (!isEmpty(entry.blobUrl) && std::strcmp(entry.blobUrl, "(failed)") != 0) {
      JS_revokeBlobUrl(entry.blobUrl);
    }
  }
  assetBrowserEntries = getAssetBrowserEntries();
}


static const char *getBlobUrl(AssetBrowserEntry &entry) {
  if (!isEmpty(entry.blobUrl)) {
    return entry.blobUrl;
  }
  if (Scoped url { JS_getAssetBlobUrl(entry.name) }) {
    copy(entry.blobUrl, url);
  } else {
    copy(entry.blobUrl, "(failed)");
  }
  return entry.blobUrl;
}

static const char *getBlobUrl(const char *assetName) {
  auto assetNameHash = hash(assetName);
  for (auto &entry : assetBrowserEntries) {
    if (entry.nameHash == assetNameHash && !std::strcmp(entry.name, assetName)) {
      return getBlobUrl(entry);
    }
  }
  return "(failed)";
}


struct AssetBrowserParams {
  bool show = false;
  const char *suffix = "";
};
void uiAssetBrowser(AssetBrowserParams params, Invocable<const char *> auto &&select) {
  static auto activeToken = -1;
  auto token = uiGetToken();
  if (params.show && activeToken != token) {
    reloadAssetBrowserEntries();
    activeToken = token;
  }
  if (activeToken == token) {
    ui("div")("asset-browser-background")("click", [&]() {
      activeToken = -1;
    })([&]() {
      ui("div")("asset-browser")([&]() {
        ui("div")("content")([&]() {
          auto suffixLen = std::strlen(params.suffix);
          for (auto &entry : assetBrowserEntries) {
            if (suffixLen > 0) {
              if (auto nameLen = std::strlen(entry.name); nameLen < suffixLen
                  || std::strcmp(&entry.name[nameLen - suffixLen], params.suffix) != 0) {
                continue;
              }
            }
            ui("div", entry.name)("cell")("click", [&]() {
              select(entry.name);
              activeToken = -1;
            })([&]() {
              if (endsWith(entry.name, ".png")) {
                ui("div")("thumbnail-container")([&]() {
                  ui("img")("thumbnail checker")("src", getBlobUrl(entry));
                });
              }
              ui("div")("filename")([&]() {
                uiText(entry.name);
              });
            });
          }
        });
      });
    });
  }
}


JS_DEFINE(char *, JS_showSaveAssetDialog, (const char *defaultName, const char *extension), {
  if (electron) {
    const win = electron.remote.getCurrentWindow();
    const options = {};
    if (defaultName) {
      options.defaultPath = assetsPath + "/" + UTF8ToString(defaultName);
    } else {
      options.defaultPath = assetsPath;
    }
    options.filters = [];
    options.filters.push({});
    const extensionStr = UTF8ToString(extension);
    options.filters[0].name = extensionStr;
    options.filters[0].extensions = [extensionStr];
    options.showsTagField = false;
    const path = electron.remote.dialog.showSaveDialogSync(win, options);
    if (path) {
      const assetName = path.replace(new RegExp("^.*[\\\\\\\\\/]"), "");
      return allocate(intArrayFromString(assetName), ALLOC_NORMAL);
    }
  } else {
    alert("Cannot save assets from the browser.");
  }
  return 0;
});


//
// Toolbar
//

static void uiEditToolbar() {
  // Play / stop
  ui("button")(edit.enabled ? "play" : "stop")(
      "title", edit.enabled ? "play game (spacebar)" : "stop game (spacebar)")(
      "click", rl::KEY_SPACE, [&]() {
        if (edit.enabled) {
          playEdit();
        } else {
          stopEdit();
        }
      });
  if (!edit.enabled) {
    return;
  }
  ui("div")("small-gap");

  // Add
  ui("div")([&]() {
    auto show = false;
    ui("button")("add-entity")("title", "add (a)")("click", rl::KEY_A, [&]() {
      show = true;
    });
    uiAssetBrowser({ .show = show, .suffix = ".bp" }, [&](const char *blueprintName) {
      clear<EditSelect>();
      auto newEnt = readBlueprint(blueprintName);
      add<EditSelect>(newEnt);
      add<EditMove>(newEnt, { .delta { edit.camera.target } });
      applyEditMoves();
      saveEditSnapshot("add");
    });
  });
  ui("div")("flex-gap");

  // Select
  ui("button")("select")("selected", isEditMode("select"))("title", "select (s)")(
      "click", rl::KEY_S, [&]() {
        setEditMode("select");
      });
  ui("div")("small-gap");
  auto hasSelection = false;
  each([&](Entity ent, EditSelect &sel) {
    hasSelection = true;
  });
  if (!hasSelection && !(isEditMode("select") || isEditMode("camera pan"))) {
    setEditMode("select");
  }

  // Move
  ui("button")("move")("disabled", !hasSelection)("selected", isEditMode("move"))(
      "title", "move (g)")("click", rl::KEY_G, [&]() {
    setEditMode("move");
  });
  ui("div")("small-gap");

  // Duplicate
  ui("button")("duplicate")("disabled", !hasSelection)("title", "duplicate (shift+d)")(
      "click", rl::KEY_LEFT_SHIFT, rl::KEY_D, [&]() {
        each([&](Entity ent, EditSelect &sel) {
          Scoped bp { writeBlueprint(ent) };
          auto newEnt = readBlueprint(bp);
          remove<EditSelect>(ent);
          add<EditSelect>(newEnt);
          add<EditMove>(newEnt, { .delta { 32, 32 } });
        });
        applyEditMoves();
        saveEditSnapshot("duplicate");
      });

  // Delete
  ui("button")("delete")("disabled", !hasSelection)("title", "delete (backspace)")(
      "click", rl::KEY_BACKSPACE, [&]() {
        each([&](Entity ent, EditSelect &sel) {
          add<EditDelete>(ent);
        });
        saveEditSnapshot("delete");
        each([&](Entity ent, EditDelete &del) {
          destroyEntity(ent);
        });
      });
  ui("div")("flex-gap");

  // Undo / redo
  ui("button")("undo")("disabled", !canUndoEdit())("title", "undo (ctrl+z)")(
      "click", rl::KEY_LEFT_CONTROL, rl::KEY_Z, [&]() {
        undoEdit();
      });
  ui("button")("redo")("disabled", !canRedoEdit())("title", "redo (ctrl+y)")(
      "click", rl::KEY_LEFT_CONTROL, rl::KEY_Y, [&]() {
        redoEdit();
      });
  ui("div")("small-gap");

  // Pan
  ui("button")("pan")("selected", isEditMode("camera pan"))("title", "camera pan (v)")(
      "click", rl::KEY_V, [&]() {
        setEditMode("camera pan");
      });

  // Zoom
  ui("button")("zoom-in")("title", "zoom in (= or scroll up)")("click", rl::KEY_EQUAL, [&]() {
    setEditZoomLevel(edit.zoomLevel + 1);
  });
  ui("button")("zoom-out")("title", "zoom out (- or scroll down)")("click", rl::KEY_MINUS, [&]() {
    setEditZoomLevel(edit.zoomLevel - 1);
  });
  ui("div")("small-gap");

  // Open
  ui("div")([&]() {
    auto show = false;
    ui("button")("open")("title", "open (ctrl+o)")("click", rl::KEY_LEFT_CONTROL, rl::KEY_O, [&]() {
      show = true;
    });
    uiAssetBrowser({ .show = show, .suffix = ".scn" }, [&](const char *sceneName) {
      openSceneEdit(sceneName);
    });
  });

  // Save
  ui("button")("save")("title", "save (ctrl+s)")("click", rl::KEY_LEFT_CONTROL, rl::KEY_S, [&]() {
    if (Scoped sceneName { JS_showSaveAssetDialog(edit.sceneName, "scn") }) {
      saveSceneEdit(sceneName);
    }
  });
}


//
// Inspect
//

template<typename T, typename PropTag>
static constexpr PropAttribs typeAttribs;

// Implements `void inspect(T *, EditInspectContext *)`
template<typename T>
concept IsVoid = std::is_same_v<T, void>;
template<typename T>
concept InspectPtr = !std::is_pointer_v<T> && requires(T * val, EditInspectContext &ctx) {
  { inspect(val, ctx) } -> IsVoid;
};
static void inspect(InspectPtr auto &val, EditInspectContext &ctx) {
  inspect(&val, ctx);
}

// int
static void inspect(int &val, EditInspectContext &ctx) {
  static char str[64];
  bprint(str, "%d", val);
  ui("input")("type", "number")("step", "any")("value", str)("change", [&](const char *newStr) {
    auto oldVal = val;
    std::sscanf(newStr, "%d", &val);
    ctx.changed = val != oldVal;
    return bprint(str, "%d", val);
  });
}

// double
static void inspect(double &val, EditInspectContext &ctx) {
  static char str[64];
  const auto printToStr = [&]() {
    bprint(str, "%.4f", val);
    for (auto i = int(std::strlen(str)) - 1; i > 0; --i) {
      if (str[i] != '0') {
        if (str[i] == '.') {
          --i;
        }
        str[i + 1] = '\0';
        break;
      }
    }
  };
  printToStr();
  ui("input")("type", "number")("step", "any")("value", str)("change", [&](const char *newStr) {
    auto oldVal = val;
    std::sscanf(newStr, "%lf", &val);
    ctx.changed = val != oldVal;
    printToStr();
    return str;
  });
}

// float
static void inspect(float &val, EditInspectContext &ctx) {
  auto d = double(val);
  inspect(d, ctx);
  val = float(d);
}

// bool
static void inspect(bool &val, EditInspectContext &ctx) {
  ui("button")(val ? "on" : "off")("click", [&]() {
    val = !val;
    ctx.changed = true;
  });
}

// String common
void inspect(const char *val, EditInspectContext &ctx,
    InvocableR<const char *, const char *> auto &&change) {
  if (ctx.attribs.asset) {
    ui("div")([&]() {
      if (endsWith(val, ".png")) {
        ui("img")("asset-preview checker")("src", getBlobUrl(val));
      }
      auto show = false;
      ui("button")("show-asset-browser")("label", val)("click", [&]() {
        show = true;
      });
      auto suffix = ctx.attribs.image ? ".png" : ctx.attribs.shader ? ".frag" : "";
      uiAssetBrowser({ .show = show, .suffix = suffix }, [&](const char *newVal) {
        ctx.changed = true;
        change(newVal);
      });
    });
  } else {
    auto tag = ctx.attribs.multiline ? "textarea" : "input";
    ui(tag)("value", val)("change", [&](const char *newVal) {
      if (std::strcmp(val, newVal) != 0) {
        ctx.changed = true;
        return change(newVal);
      }
      return val;
    });
  }
}

// char array
template<int N, typename PropTag>
static constexpr PropAttribs typeAttribs<char[N], PropTag> = {
  .breakBefore = PropTag::attribs.asset,
  .breakAfter = PropTag::attribs.asset,
};
template<int N>
static void inspect(char (&val)[N], EditInspectContext &ctx) {
  inspect(val, ctx, [&](const char *newVal) {
    copy(val, newVal);
    return (const char *)val;
  });
}

// gx::String
template<typename PropTag>
static constexpr PropAttribs typeAttribs<gx::String, PropTag> = {
  .breakBefore = PropTag::attribs.asset,
  .breakAfter = PropTag::attribs.asset,
};
static void inspect(gx::String &val, EditInspectContext &ctx) {
  inspect(val, ctx, [&](const char *newVal) {
    val = newVal;
    return (const char *)val;
  });
}

// Vec2
static void inspect(Vec2 &val, EditInspectContext &ctx) {
  ui("div")("numbers-row")([&]() {
    inspect(val.x, ctx);
    inspect(val.y, ctx);
  });
}

// rl::Rectangle
static void inspect(rl::Rectangle &val, EditInspectContext &ctx) {
  ui("div")("numbers-row")([&]() {
    ui("div")("numbers-row")([&]() {
      inspect(val.x, ctx);
      inspect(val.y, ctx);
    });
    ui("div")("numbers-row")([&]() {
      inspect(val.width, ctx);
      inspect(val.height, ctx);
    });
  });
}

// gx::Slice<T>
template<typename T, typename PropTag>
static constexpr PropAttribs typeAttribs<gx::Slice<T>, PropTag> {
  .breakBefore = true,
  .breakAfter = true,
};
static auto inspectSliceActiveMenuToken = -1;
template<typename T>
static void inspect(gx::Slice<T> &val, EditInspectContext &ctx) {
  auto nElems = len(val);
  auto removeI = -1, swapNextI = -1, addBeforeI = -1;
  auto i = 0;
  for (auto &elem : val) {
    ui("div")("seq-elem-container")([&]() {
      auto token = uiGetToken();
      ui("div")([&]() {
        ui("button")("show-seq-elem-menu")("click", [&]() {
          inspectSliceActiveMenuToken = token;
        });
      });
      ui("div")("seq-elem-menu-anchor")([&]() {
        inspect(elem, ctx);
        if (inspectSliceActiveMenuToken == token) {
          ui("div")("seq-elem-menu-background")("click", [&]() {
            inspectSliceActiveMenuToken = -1;
          });
          ui("div")("seq-elem-menu-container")([&]() {
            ui("button")("remove")("label", "remove item")("click", [&]() {
              removeI = i;
            });
            if (i > 0) {
              ui("button")("up")("label", "move up")("click", [&]() {
                swapNextI = i - 1;
              });
            }
            if (i < nElems - 1) {
              ui("button")("down")("label", "move down")("click", [&]() {
                swapNextI = i;
              });
            }
            ui("button")("add")("label", "add before")("click", [&]() {
              addBeforeI = i;
            });
          });
        }
      });
    });
    ++i;
  }
  if (removeI >= 0) {
    remove(val, removeI);
    ctx.changed = true;
    inspectSliceActiveMenuToken = -1;
  }
  if (swapNextI >= 0) {
    std::swap(val[swapNextI], val[swapNextI + 1]);
    ctx.changed = true;
    inspectSliceActiveMenuToken = -1;
  }
  if (addBeforeI >= 0) {
    insert(val, addBeforeI, T {});
    ctx.changed = true;
    inspectSliceActiveMenuToken = -1;
  }
  ui("div")("seq-elem-container")([&]() {
    ui("div")([&]() {
      ui("button")("show-seq-elem-menu")("disabled", true);
    });
    ui("button")("add-seq-elem")("click", [&]() {
      append(val);
      ctx.changed = true;
    });
  });
}

// Props
template<Props T>
static void inspect(T &val, EditInspectContext &ctx) {
  ui("div")("props-container")([&]() {
    forEachProp(val, [&]<typename PropTag, typename Prop>(PropTag propTag, Prop &propVal) {
      constexpr auto &attribs = propTag.attribs;
      constexpr auto name = attribs.name;
      if constexpr (attribs.breakBefore || typeAttribs<Prop, PropTag>.breakBefore) {
        ui("div")("prop-break");
      }
      ui("div")("prop-container")([&]() {
        ui("div")("prop-name")([&]() {
          uiText(name.data());
        });
        ui("div")("prop-value-container")([&]() {
          auto prevChanged = ctx.changed;
          auto prevAttribs = ctx.attribs;
          ctx.attribs = attribs;
          if constexpr (requires { inspect(propTag, &val, &ctx); }) {
            inspect(propTag, &val, &ctx);
          } else {
            inspect(propVal, ctx);
          }
          ctx.attribs = prevAttribs;
          if (ctx.changed && !prevChanged) {
            if constexpr (isComponentType<T> && requires { change(propTag, &val, ctx.ent); }) {
              change(propTag, &val, ctx.ent);
            }
            if (isEmpty(ctx.changeDescription)) {
              bprint(ctx.changeDescription, "change %s %s", ctx.componentTitle, name.data());
            }
          }
        });
      });
      if constexpr (attribs.breakAfter || typeAttribs<Prop, PropTag>.breakAfter) {
        ui("div")("prop-break");
      }
    });
  });
}

// Inspector
static void uiEditInspector() {
  static const auto titleify = [&](std::string_view title) {
    static char buf[64];
    auto i = 0;
    for (auto c : title) {
      if (i >= int(sizeof(buf) - 2)) {
        break;
      }
      if (std::isupper(c)) {
        if (i > 0) {
          buf[i++] = ' ';
        }
        buf[i++] = char(std::tolower(c));
      } else {
        buf[i++] = c;
      }
    }
    buf[i] = '\0';
    return buf;
  };

  auto first = true;
  each([&](Entity ent, EditSelect &sel) {
    if (!first) {
      return;
    }
    first = false;

    char nextInspectedComponentTitle[sizeof(edit.inspectedComponentTitle)] = "";
    void (*remover)(Entity ent) = nullptr;

    // Section for each component entity has
    forEachComponentType([&]<typename T>() {
      if (has<T>(ent)) {
        auto title = titleify(getTypeName<T>());
        auto open = !std::strcmp(edit.inspectedComponentTitle, title);
        ui("details", title)(title)("open", true)([&]() {
          // Title
          ui("summary")("click", [&]() {
            if (open) {
              clear(edit.inspectedComponentTitle);
            } else {
              copy(nextInspectedComponentTitle, title);
            }
          })([&]() {
            uiText(title);
            ui("button")("remove")("click", [&]() {
              remover = [](Entity ent) {
                remove<T>(ent);
                auto title = titleify(getTypeName<T>());
                copy(edit.inspectedComponentTitle, title);
                saveEditSnapshot(bprint<64>("remove %s", title), true);
              };
            });
          });

          // Properties
          EditInspectContext ctx { .ent = ent, .componentTitle = title };
          auto &comp = get<T>(ent);
          ui("div")("component-container")([&]() {
            inspect(comp, ctx);
          });
          if (!isEmpty(ctx.changeDescription)) {
            saveEditSnapshot(ctx.changeDescription, true);
          } else if (ctx.changed) {
            saveEditSnapshot(bprint<64>("edit %s", title), true);
          }
        });
      }
    });

    // Buttons to add components entity doesn't have
    ui("div")("button-bar")([&]() {
      forEachComponentType([&]<typename T>() {
        if (!has<T>(ent)) {
          auto title = titleify(getTypeName<T>());
          ui("button")("add")("label", title)("click", [&]() {
            add<T>(ent);
            copy(edit.inspectedComponentTitle, title);
            saveEditSnapshot(bprint<64>("add %s", title), true);
          });
        }
      });
    });

    // Blueprint actions
    ui("div")("button-bar")([&]() {
      ui("button")("save")("label", "save blueprint")("title", "ctrl+b")(
          "click", rl::KEY_LEFT_CONTROL, rl::KEY_B, [&]() {
            if (Scoped assetName { JS_showSaveAssetDialog(nullptr, "bp") }) {
              Vec2 pos { 0, 0 };
              if (has<EditBox>(ent)) {
                auto &box = get<EditBox>(ent);
                pos = { box.rect.x + 0.5f * box.rect.width, box.rect.y + 0.5f * box.rect.height };
                add<EditMove>(ent, { .delta { -pos } });
                applyEditMoves();
              }
              Scoped jsn { writeBlueprint(ent) };
              writeAssetContents(assetName, stringify(jsn, true));
              if (pos != Vec2 { 0, 0 }) {
                add<EditMove>(ent, { .delta { pos } });
                applyEditMoves();
              }
            }
          });
    });

    // Doing these after UI calls to prevent rendering insconsistent state
    if (!isEmpty(nextInspectedComponentTitle)) {
      copy(edit.inspectedComponentTitle, nextInspectedComponentTitle);
    }
    if (remover) {
      remover(ent);
    }
  });
}


//
// Status
//

static void uiEditStatus() {
  // FPS
  ui("div")([&]() {
    uiText("fps: %d", rl::GetFPS());
  });

  ui("div")("flex-gap");

  // Notification
  if (!isEmpty(edit.notification)) {
    if (rl::GetTime() - edit.lastNotificationTime < 3.5) {
      ui("div")([&]() {
        uiText(edit.notification);
      });
    } else {
      clear(edit.notification);
    }
  }

  if (edit.enabled) {
    ui("div")("small-gap");

    // Mode
    ui("div")([&]() {
      uiText(edit.mode);
    });
  }
}


//
// Top-level
//

JS_DEFINE(void, JS_setCanvasModeClass, (const char *prevMode, const char *mode), {
  const el = document.getElementById("canvas");
  if (el) {
    const prevModeStr = "mode-" + UTF8ToString(prevMode).replace(new RegExp(" ", "g"), "-");
    if (prevModeStr.length > 0) {
      el.classList.remove(prevModeStr);
    }
    const modeStr = "mode-" + UTF8ToString(mode).replace(new RegExp(" ", "g"), "-");
    el.classList.add(modeStr);
  }
})

void uiEdit() {
  {
#ifdef __EMSCRIPTEN__
    auto mode = edit.enabled ? edit.mode : "select";
    static char prevMode[sizeof(edit.mode)] = "";
    if (std::strcmp(mode, prevMode) != 0) {
      JS_setCanvasModeClass(prevMode, mode);
      copy(prevMode, mode);
    }
#endif
  }

  uiPatch("top", [&]() {
    ui("div")("toolbar")([&]() {
      uiEditToolbar();
    });
  });
  uiPatch("side", [&]() {
    ui("div")("inspector")([&]() {
      uiEditInspector();
    });
  });
  uiPatch("bottom", [&]() {
    ui("div")("status")([&]() {
      uiEditStatus();
    });
  });
}
