#pragma once
#include "core/core.hh"


//
// Invocable
//

template<typename T, typename... Args>
concept Invocable = std::is_invocable_v<T, Args...>;

template<typename T, typename Result, typename... Args>
concept InvocableR = std::is_invocable_r_v<Result, T, Args...>;


//
// Type names
//

template<typename T>
constexpr std::string_view getTypeName() {
  constexpr auto prefixLength = 36, suffixLength = 1;
  const char *data = __PRETTY_FUNCTION__;
  auto end = data;
  while (*end) {
    ++end;
  }
  return { data + prefixLength, size_t(end - data - prefixLength - suffixLength) };
}


//
// Component types list
//

template<int N>
struct ComponentTypeCounter : ComponentTypeCounter<N - 1> {
  static constexpr auto num = N;
};
template<>
struct ComponentTypeCounter<0> {
  static constexpr auto num = 0;
};
ComponentTypeCounter<0> numComponentTypes(ComponentTypeCounter<0>);

template<int I>
struct ComponentTypeList;
template<>
struct ComponentTypeList<0> {
  static void each(auto &&f) {
  }
};

inline constexpr auto maxNumComponentTypes = 32;

template<typename T>
inline constexpr auto isComponentType = false;

#define ComponentTypeListAdd(T)                                                                    \
  template<>                                                                                       \
  inline constexpr auto isComponentType<T> = true;                                                 \
  constexpr auto ComponentTypeList_##T##_Size                                                      \
      = decltype(numComponentTypes(ComponentTypeCounter<maxNumComponentTypes>()))::num + 1;        \
  static_assert(ComponentTypeList_##T##_Size < maxNumComponentTypes);                              \
  ComponentTypeCounter<ComponentTypeList_##T##_Size> numComponentTypes(                            \
      ComponentTypeCounter<ComponentTypeList_##T##_Size>);                                         \
  template<>                                                                                       \
  struct ComponentTypeList<ComponentTypeList_##T##_Size> {                                         \
    static void each(auto &&f) {                                                                   \
      ComponentTypeList<ComponentTypeList_##T##_Size - 1>::each(f);                                \
      f.template operator()<T>();                                                                  \
    }                                                                                              \
  }

#define Comp(T)                                                                                    \
  T;                                                                                               \
  ComponentTypeListAdd(T);                                                                         \
  struct T

#define UseComponentTypes()                                                                        \
  static void forEachComponentType(auto &&f) {                                                     \
    ComponentTypeList<decltype(numComponentTypes(                                                  \
        ComponentTypeCounter<maxNumComponentTypes>()))::num>::each(std::forward<decltype(f)>(f));  \
  }


//
// Props
//

constexpr uint32_t hash(std::string_view str);

struct PropAttribs {
  std::string_view name;
  uint32_t nameHash = hash(name);

  bool breakBefore = false;
  bool breakAfter = false;
  bool multiline = false;
  bool asset = false;
  bool image = false;
  bool shader = false;
};

inline constexpr auto maxNumProps = 24;

template<int N>
struct PropCounter : PropCounter<N - 1> {
  static constexpr auto num = N;
};
template<>
struct PropCounter<0> {
  static constexpr auto num = 0;
};
static PropCounter<0> numProps(PropCounter<0>);

template<int N>
struct PropIndex {
  static constexpr auto index = N;
};

template<typename T, int N>
struct PropTagWrapper {
  struct Tag {
    inline static constexpr PropAttribs attribs = T::getPropAttribs(PropIndex<N> {});
  };
};
template<typename T, int N>
struct PropTagWrapper<const T, N> {
  using Tag = typename PropTagWrapper<T, N>::Tag;
};
template<typename T, int N>
using PropTag = typename PropTagWrapper<T, N>::Tag;

#define Prop(type, name_, ...) PropNamed(#name_, type, name_, __VA_ARGS__)
#define PropNamed(nameStr, type, name_, ...)                                                       \
  using name_##_Index = PropIndex<decltype(numProps(PropCounter<maxNumProps>()))::num>;            \
  static PropCounter<decltype(numProps(PropCounter<maxNumProps>()))::num + 1> numProps(            \
      PropCounter<decltype(numProps(PropCounter<maxNumProps>()))::num + 1>);                       \
  static std::type_identity<PROP_PARENS_1(PROP_PARENS_3 type)> propType(name_##_Index);            \
  static constexpr PropAttribs getPropAttribs(name_##_Index) {                                     \
    return { .name = #name_, __VA_ARGS__ };                                                        \
  };                                                                                               \
  std::type_identity_t<PROP_PARENS_1(PROP_PARENS_3 type)> name_
#define PROP_PARENS_1(...) PROP_PARENS_2(__VA_ARGS__)
#define PROP_PARENS_2(...) NO##__VA_ARGS__
#define PROP_PARENS_3(...) PROP_PARENS_3 __VA_ARGS__
#define NOPROP_PARENS_3

template<auto memPtr>
struct ContainingType {};
template<typename C, typename R, R C::*memPtr>
struct ContainingType<memPtr> {
  using Type = C;
};
#define PropTag(field) PropTag<ContainingType<&field>::Type, field##_Index::index>

struct CanConvertToAnything {
  template<typename Type>
  operator Type() const; // NOLINT(google-explicit-constructor)
};
template<typename Aggregate, typename Base = std::index_sequence<>, typename = void>
struct CountFields : Base {};
template<typename Aggregate, int... Indices>
struct CountFields<Aggregate, std::index_sequence<Indices...>,
    std::void_t<decltype(
        Aggregate { { (static_cast<void>(Indices), std::declval<CanConvertToAnything>()) }...,
            { std::declval<CanConvertToAnything>() } })>>
    : CountFields<Aggregate, std::index_sequence<Indices..., sizeof...(Indices)>> {};
template<typename T>
constexpr int countFields() {
  return CountFields<std::remove_cvref_t<T>>().size();
}

template<typename T>
concept Props = std::is_aggregate_v<T>;

template<Props T, typename F>
inline void forEachProp(T &val, F &&func) {
  if constexpr (requires { forEachField(const_cast<std::remove_cvref_t<T> &>(val), func); }) {
    forEachField(const_cast<std::remove_cvref_t<T> &>(val), func);
  } else if constexpr (requires { T::propType(PropIndex<0> {}); }) {
    constexpr auto n = countFields<T>();
    const auto call = [&]<typename Index>(Index index, auto &val) {
      if constexpr (requires { T::propType(index); }) {
        static_assert(std::is_same_v<typename decltype(T::propType(index))::type,
            std::remove_cvref_t<decltype(val)>>);
        func(PropTag<T, Index::index> {}, val);
      }
    };
#define C(i) call(PropIndex<i> {}, f##i)
    if constexpr (n == 1) {
      auto &[f0] = val;
      (C(0));
    } else if constexpr (n == 2) {
      auto &[f0, f1] = val;
      (C(0), C(1));
    } else if constexpr (n == 3) {
      auto &[f0, f1, f2] = val;
      (C(0), C(1), C(2));
    } else if constexpr (n == 4) {
      auto &[f0, f1, f2, f3] = val;
      (C(0), C(1), C(2), C(3));
    } else if constexpr (n == 5) {
      auto &[f0, f1, f2, f3, f4] = val;
      (C(0), C(1), C(2), C(3), C(4));
    } else if constexpr (n == 6) {
      auto &[f0, f1, f2, f3, f4, f5] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5));
    } else if constexpr (n == 7) {
      auto &[f0, f1, f2, f3, f4, f5, f6] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6));
    } else if constexpr (n == 8) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7));
    } else if constexpr (n == 9) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8));
    } else if constexpr (n == 10) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9));
    } else if constexpr (n == 11) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10));
    } else if constexpr (n == 12) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11));
    } else if constexpr (n == 13) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12));
    } else if constexpr (n == 14) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13));
    } else if constexpr (n == 15) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14));
    } else if constexpr (n == 16) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15));
    } else if constexpr (n == 17) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16));
    } else if constexpr (n == 18) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17] = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17));
    } else if constexpr (n == 19) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18));
    } else if constexpr (n == 20) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18,
          f19]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18), C(19));
    } else if constexpr (n == 21) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18,
          f19, f20]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18), C(19), C(20));
    } else if constexpr (n == 22) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18,
          f19, f20, f21]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18), C(19), C(20), C(21));
    } else if constexpr (n == 23) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18,
          f19, f20, f21, f22]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18), C(19), C(20), C(21), C(22));
    } else if constexpr (n == 24) {
      auto &[f0, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15, f16, f17, f18,
          f19, f20, f21, f22, f23]
          = val;
      (C(0), C(1), C(2), C(3), C(4), C(5), C(6), C(7), C(8), C(9), C(10), C(11), C(12), C(13),
          C(14), C(15), C(16), C(17), C(18), C(19), C(20), C(21), C(22), C(23));
    }
#undef C
  }
}
