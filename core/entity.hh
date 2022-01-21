#pragma once
#include "core/core.hh"


//
// Entities
//

struct Entity {
  entt::entity val = entt::null;
  Entity() = default;
  Entity(entt::entity val_) // NOLINT(google-explicit-constructor)
      : val(val_) {
  }
  explicit Entity(uint32_t i)
      : val(entt::entity(i)) {
  }
  operator entt::entity() const { // NOLINT(google-explicit-constructor)
    return val;
  }
  explicit operator uint32_t() const {
    return entt::to_integral(val);
  }
};
constexpr Entity nullEntity {};


// The global entity registry
inline entt::registry registry;

// Create a new entity
template<typename... Ts>
inline Entity createEntity(Ts... values) {
  auto ent = registry.create();
  (add<Ts>(ent, std::move(values)), ...);
  return ent;
}
template<typename... Ts>
inline Entity createEntity(Entity hint, Ts... values) {
  auto ent = registry.create(hint);
  (add<Ts>(ent, std::move(values)), ...);
  return ent;
}
struct CreateEntity {
  Entity ent;
  template<typename... Ts>
  CreateEntity(Ts... values) // NOLINT(google-explicit-constructor)
      : ent(createEntity(std::move(values)...)) {
  }
};

// Destroy an entity
inline Seq<void (*)(Entity), 32> componentRemovers;
inline void destroyEntity(Entity ent) {
  for (auto &remover : componentRemovers) {
    remover(ent);
  }
  registry.destroy(ent);
}

// Check that an entity exists
inline bool exists(Entity ent) {
  return registry.valid(ent);
}

// Get the number of entities that currently exist
inline int numEntities() {
  return int(registry.alive());
}


//
// Components
//

// Component pools
template<typename T>
inline void remove(Entity ent);
template<typename F>
void each(F &&f);
inline bool entitiesClearedOnExit = false;
template<typename T>
struct ComponentPool {
  Seq<int> sparse;
  struct DenseElem {
    Entity ent;
    T *value;
  };
  Seq<DenseElem> dense;

  ComponentPool() {
    append(componentRemovers, remove<T>);
  }

  ~ComponentPool() {
    if (!entitiesClearedOnExit) {
      each(destroyEntity);
      entitiesClearedOnExit = true;
    }
  }

  static int sparseIndex(Entity ent) {
    return int(uint32_t(ent) & uint32_t(0xFFFFF));
  }

  bool contains(Entity ent) const {
    auto sparseI = sparseIndex(ent);
    return sparseI < len(sparse) && sparse[sparseI] != -1;
  }

  T &get(Entity ent) {
    return *dense[sparse[sparseIndex(ent)]].value;
  }

  T &emplace(Entity ent, T value) {
    auto sparseI = sparseIndex(ent);
    if (sparseI >= sparse.capacity) {
      auto oldCapacity = sparse.capacity;
      sparse.capacity = sparseI == 0 ? 32 : sparseI << 1;
      sparse.size = sparse.capacity;
      sparse.data = (int *)std::realloc(sparse.data, sizeof(int) * sparse.capacity);
      auto newCapacity = sparse.capacity;
      for (auto i = oldCapacity; i < newCapacity; ++i) {
        sparse.data[i] = -1;
      }
    }
    append(dense, { ent, new T(std::move(value)) });
    auto maxDenseI = len(dense) - 1;
    sparse[sparseI] = maxDenseI;
    return *dense[maxDenseI].value;
  }

  void erase(Entity ent) {
    auto sparseI = sparseIndex(ent);
    auto denseI = sparse[sparseI];
    sparse[sparseI] = -1;
    delete dense[denseI].value;
    auto maxDenseI = len(dense) - 1;
    if (denseI < maxDenseI) {
      std::memcpy(&dense[denseI], &dense[maxDenseI], sizeof(DenseElem));
      sparse[sparseIndex(dense[denseI].ent)] = denseI;
    }
    --dense.size;
  }
};
template<typename T>
inline ComponentPool<T> componentPool;

// Check whether entity has a component
template<typename T>
inline bool has(Entity ent) {
  return exists(ent) && componentPool<T>.contains(ent);
}

// Get a component on an entity
template<typename T>
inline decltype(auto) get(Entity ent) {
  return componentPool<T>.get(ent);
}
template<typename T>
inline T *getPtr(Entity ent) {
  return has<T>(ent) ? &componentPool<T>.get(ent) : nullptr;
}

// Add a component to an entity
template<typename T>
inline decltype(auto) add(Entity ent, T value = {}) {
  if (has<T>(ent)) {
    return get<T>(ent);
  } else {
    auto &comp = componentPool<T>.emplace(ent, std::move(value));
    if constexpr (requires { add(&comp, ent); }) {
      add(&comp, ent);
    }
    if constexpr (isComponentType<T>) {
      forEachProp(comp, [&](auto propTag, auto &propVal) {
        if constexpr (requires { change(propTag, &comp, ent); }) {
          change(propTag, &comp, ent);
        }
      });
    }
    return comp;
  }
}
template<typename T>
inline T *addPtr(Entity ent, T value = {}) {
  return &add(ent, std::move(value));
}

// Remove a component from an entity
template<typename T>
inline void remove(Entity ent) {
  if (has<T>(ent)) {
    if constexpr (requires { remove((T *)nullptr, ent); }) {
      remove(&get<T>(ent), ent);
    }
    componentPool<T>.erase(ent);
  }
}

// Query component combinations
template<typename T>
struct Each {};
template<typename R, typename C>
struct Each<R (C::*)(Entity ent) const> {
  template<typename F>
  static void each(F &&f) {
    registry.each(std::forward<F>(f));
  }
};
template<typename R, typename C, typename T, typename... Ts>
struct Each<R (C::*)(Entity ent, T &, Ts &...) const> {
  static void each(auto &&f) {
    for (int i = componentPool<T>.dense.size - 1; i >= 0; --i) {
      auto &elem = componentPool<T>.dense[i];
      auto ent = elem.ent;
      if ((componentPool<Ts>.contains(ent) && ...)) {
        f(ent, *elem.value, componentPool<Ts>.get(ent)...);
      }
    }
  }
};
template<typename R, typename C, typename T, typename... Ts>
struct Each<R (C::*)(Entity ent, T *, Ts *...) const> {
  static void each(auto &&f) {
    for (int i = componentPool<T>.dense.size - 1; i >= 0; --i) {
      auto &elem = componentPool<T>.dense[i];
      auto ent = elem.ent;
      if ((componentPool<Ts>.contains(ent) && ...)) {
        f(ent, elem.value, &componentPool<Ts>.get(ent)...);
      }
    }
  }
};
template<typename R, typename... Args>
inline void each(R (&f)(Args...)) {
  each([&](Args... args) {
    f(std::forward<Args>(args)...);
  });
}
template<typename F>
inline void each(F &&f) {
  if constexpr (requires { &F::operator(); }) {
    Each<decltype(&F::operator())>::each(std::forward<F>(f));
  } else {
    each(std::forward<F>(f));
  }
}

// Remove a component on all entities
template<typename T>
inline void clear() {
  each([&](Entity ent, T &comp) {
    if constexpr (requires { remove(&comp, ent); }) {
      remove(&comp, ent);
    }
    componentPool<T>.erase(ent);
  });
}

// Sort component data
template<typename T, typename F>
inline void sort(F &&f) {
  auto &pool = componentPool<T>;
  std::sort(pool.dense.begin(), pool.dense.end(), [&](auto &a, auto &b) {
    return f(*b.value, *a.value);
  });
  for (auto i = 0; auto &elem : pool.dense) {
    pool.sparse[pool.sparseIndex(elem.ent)] = i++;
  }
}
template<typename T, typename F>
inline void sortPtr(F &&f) {
  sort<T>([&](const T &a, const T &b) {
    return f(&const_cast<T &>(a), &const_cast<T &>(b));
  });
}
