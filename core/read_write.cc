#include "core.hh"

#include "core/game.hh"
UseComponentTypes();


//
// Scene
//

// Blueprint

Entity readBlueprint(const cJSON *jsn) {
  Entity ent;
  if (auto idJsn = cJSON_GetObjectItemCaseSensitive(jsn, "id")) {
    ent = createEntity(Entity(cJSON_GetNumberValue(idJsn)));
  } else {
    ent = createEntity();
  }
  for (auto compJsn = cJSON_GetObjectItemCaseSensitive(jsn, "components")->child; compJsn;
       compJsn = compJsn->next) {
    const auto typeName = cJSON_GetStringValue(cJSON_GetObjectItemCaseSensitive(compJsn, "_type"));
    const auto typeNameHash = hash(typeName);
    forEachComponentType([&]<typename T>() {
      constexpr auto TName = getTypeName<T>();
      constexpr auto TNameHash = hash(TName);
      if (typeNameHash == TNameHash && typeName == TName) {
        T comp {};
        read(comp, compJsn);
        add<T>(ent, std::move(comp));
      }
    });
  }
  return ent;
}

Entity readBlueprint(const char *assetName) {
  Scoped jsn { cJSON_Parse(getAssetContents(assetName)) };
  auto ent = readBlueprint(jsn);
  return ent;
}

cJSON *writeBlueprint(Entity ent, bool writeId) {
  auto result = cJSON_CreateObject();

  if (writeId) {
    auto idJsn = cJSON_CreateNumber(uint32_t(ent));
    cJSON_AddItemToObjectCS(result, "id", idJsn);
  }

  auto compsJsn = cJSON_CreateArray();
  cJSON_AddItemToObjectCS(result, "components", compsJsn);
  forEachComponentType([&]<typename T>() {
    if (has<T>(ent)) {
      auto &comp = get<T>(ent);
      auto compJsn = write(comp);

      cJSON_AddItemToObjectCS(
          compJsn, "_type", cJSON_CreateString(nullTerminate<64>(getTypeName<T>())));
      if (compJsn->child->prev != compJsn->child) {
        compJsn->child->prev->next = compJsn->child;
        compJsn->child = compJsn->child->prev;
        compJsn->child->prev->next = nullptr;
      }

      cJSON_AddItemToArray(compsJsn, compJsn);
    }
  });

  return result;
}


// Top-level

void readScene(const cJSON *jsn) {
  for (auto entityJsn = cJSON_GetObjectItemCaseSensitive(jsn, "entities")->child; entityJsn;
       entityJsn = entityJsn->next) {
    readBlueprint(entityJsn);
  }
}

void readScene(const char *assetName) {
  Scoped jsn { cJSON_Parse(getAssetContents(assetName)) };
  readScene(jsn);
}

cJSON *writeScene() {
  auto result = cJSON_CreateObject();
  Seq<Entity> entities;
  each([&](Entity ent) {
    append(entities, ent);
  });
  std::reverse(entities.begin(), entities.end());
  auto entitiesJsn = cJSON_CreateArray();
  cJSON_AddItemToObjectCS(result, "entities", entitiesJsn);
  for (auto ent : entities) {
    cJSON_AddItemToArray(entitiesJsn, writeBlueprint(ent, true));
  }
  return result;
}
