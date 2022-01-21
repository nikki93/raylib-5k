#pragma once
#include "core/core.hh"


namespace cp {


//
// Type aliases
//

// chipmunk/cpBB.h
using BB = cpBB;

// chipmunk/chipmunk_types.h
using HashValue = cpHashValue;
using CollisionID = cpCollisionID;

// chipmunk/cpSpatialIndex.h
using SpatialIndex = cpSpatialIndex;
using SpaceHash = cpSpaceHash;
using BBTree = cpBBTree;


//
// Function aliases
//

// chipmunk/cpSpatialIndex.h
inline cpSpaceHash *SpaceHashAlloc() {
  return cpSpaceHashAlloc();
}
inline cpSpatialIndex *SpaceHashInit(cpSpaceHash *hash, cpFloat celldim, int numcells,
    cpSpatialIndexBBFunc bbfunc, cpSpatialIndex *staticIndex) {
  return cpSpaceHashInit(hash, celldim, numcells, bbfunc, staticIndex);
}
inline cpSpatialIndex *SpaceHashNew(
    cpFloat celldim, int cells, cpSpatialIndexBBFunc bbfunc, cpSpatialIndex *staticIndex) {
  return cpSpaceHashNew(celldim, cells, bbfunc, staticIndex);
}
inline void SpaceHashResize(cpSpaceHash *hash, cpFloat celldim, int numcells) {
  return cpSpaceHashResize(hash, celldim, numcells);
}
inline cpBBTree *BBTreeAlloc() {
  return cpBBTreeAlloc();
}
inline cpSpatialIndex *BBTreeInit(
    cpBBTree *tree, cpSpatialIndexBBFunc bbfunc, cpSpatialIndex *staticIndex) {
  return cpBBTreeInit(tree, bbfunc, staticIndex);
}
inline cpSpatialIndex *BBTreeNew(cpSpatialIndexBBFunc bbfunc, cpSpatialIndex *staticIndex) {
  return cpBBTreeNew(bbfunc, staticIndex);
}
inline void BBTreeOptimize(cpSpatialIndex *index) {
  return cpBBTreeOptimize(index);
}
inline void SpatialIndexFree(cpSpatialIndex *index) {
  return cpSpatialIndexFree(index);
}
inline void SpatialIndexDestroy(cpSpatialIndex *index) {
  return cpSpatialIndexDestroy(index);
}
inline int SpatialIndexCount(cpSpatialIndex *index) {
  return cpSpatialIndexCount(index);
}
inline void SpatialIndexEach(cpSpatialIndex *index, cpSpatialIndexIteratorFunc func, void *data) {
  return cpSpatialIndexEach(index, func, data);
}
inline cpBool SpatialIndexContains(cpSpatialIndex *index, void *obj, cpHashValue hashid) {
  return cpSpatialIndexContains(index, obj, hashid);
}
inline void SpatialIndexInsert(cpSpatialIndex *index, void *obj, cpHashValue hashid) {
  return cpSpatialIndexInsert(index, obj, hashid);
}
inline void SpatialIndexRemove(cpSpatialIndex *index, void *obj, cpHashValue hashid) {
  return cpSpatialIndexRemove(index, obj, hashid);
}
inline void SpatialIndexReindex(cpSpatialIndex *index) {
  return cpSpatialIndexReindex(index);
}
inline void SpatialIndexReindexObject(cpSpatialIndex *index, void *obj, cpHashValue hashid) {
  return cpSpatialIndexReindexObject(index, obj, hashid);
}
template<typename F>
inline void SpatialIndexQuery(cpSpatialIndex *index, void *obj, cpBB bb, F &&f) {
  return cpSpatialIndexQuery(
      index, obj, bb,
      [](void *obj1, void *obj2, cpCollisionID id, void *data) {
        auto &f = *static_cast<F *>(data);
        return f(obj1, obj2, id);
      },
      &f);
}
inline void SpatialIndexSegmentQuery(cpSpatialIndex *index, void *obj, cpVect a, cpVect b,
    cpFloat t_exit, cpSpatialIndexSegmentQueryFunc func, void *data) {
}


}


//
// Spatial index wrapper
//

struct BoundingBox {
  Vec2 min;
  Vec2 max;
};

inline Entity spatialIndexId(void *id) {
  return Entity(uintptr_t(id) - 1);
}
inline void *spatialIndexId(Entity ent) {
  return (void *)(uintptr_t(uint32_t(ent)) + 1);
}

struct SpatialIndex {
  cp::SpatialIndex *index = nullptr;
  ~SpatialIndex() {
    cp::SpatialIndexFree(index);
    index = nullptr;
  }
};

template<typename T>
inline SpatialIndex newSpatialIndex() {
  return SpatialIndex {
    cp::BBTreeNew(
        [](void *obj) {
          if (auto ent = spatialIndexId(obj); has<T>(ent)) {
            auto bb = boundingBox(&get<T>(ent), ent);
            return cp::BB { bb.min.x, bb.min.y, bb.max.x, bb.max.y };
          } else {
            return cp::BB {};
          }
        },
        nullptr),
  };
}

inline void insert(SpatialIndex *index, Entity ent) {
  cp::SpatialIndexInsert(index->index, spatialIndexId(ent), cp::HashValue(spatialIndexId(ent)));
}

inline void remove(SpatialIndex *index, Entity ent) {
  cp::SpatialIndexRemove(index->index, spatialIndexId(ent), cp::HashValue(spatialIndexId(ent)));
}

inline void reindex(SpatialIndex *index, Entity ent) {
  cp::SpatialIndexReindexObject(
      index->index, spatialIndexId(ent), cp::HashValue(spatialIndexId(ent)));
}

inline void query(SpatialIndex *index, BoundingBox bb, Invocable<Entity> auto &&visit) {
  cp::SpatialIndexQuery(index->index, nullptr, { bb.min.x, bb.min.y, bb.max.x, bb.max.y },
      [&](void *obj, void *otherObj, cp::CollisionID id) {
        visit(spatialIndexId(otherObj));
        return id;
      });
}
