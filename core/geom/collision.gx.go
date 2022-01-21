package geom

import (
	. "github.com/nikki93/dream-hotel/core/entity"
)

//
// Shapes
//

//gx:extern BoundingBox
type BoundingBox struct {
	Min, Max Vec2
}

//gx:extern c2Poly
type Polygon struct {
	Count   int
	Verts   [8]Vec2
	Normals [8]Vec2 //gx:extern norms
}

//gx:extern c2MakePoly
func (p *Polygon) ConvexHull()

//
// Intersect
//

//gx:extern c2Manifold
type IntersectResult struct {
	Count         int
	Depths        [2]float64
	ContactPoints [2]Vec2 //gx:extern contact_points
	Normal        Vec2    //gx:extern n
}

//gx:extern c2r
type c2r struct { // Rotation
	c, s float64 // Cos, sin pair
}

//gx:extern c2x
type c2x struct { // Transform
	p Vec2
	r c2r
}

//gx:extern c2PolytoPolyManifold
func c2PolytoPolyManifold(a *Polygon, ax *c2x, b *Polygon, bx *c2x, m *IntersectResult)

func IntersectPolygons(a *Polygon, aPos Vec2, b *Polygon, bPos Vec2) IntersectResult {
	result := IntersectResult{}
	ax := c2x{aPos, c2r{1, 0}}
	bx := c2x{bPos, c2r{1, 0}}
	c2PolytoPolyManifold(a, &ax, b, &bx, &result)
	return result
}

//
// SpatialIndex
//

//gx:extern SpatialIndex
type SpatialIndex struct{}

//gx:extern newSpatialIndex
func NewSpatialIndex[T any]() (s SpatialIndex) { return }

//gx:extern insert
func (s *SpatialIndex) Insert(ent Entity)

//gx:extern remove
func (s *SpatialIndex) Remove(ent Entity)

//gx:extern reindex
func (s *SpatialIndex) Reindex(ent Entity)

//gx:extern query
func (s *SpatialIndex) Query(bb BoundingBox, visit func(ent Entity))
