package geom

import (
	. "github.com/nikki93/raylib-5k/core/entity"
)

//
// Shapes
//

//gx:extern BoundingBox
type BoundingBox struct {
	Min, Max Vec2
}

//gx:extern c2Circle
type Circle struct {
	Pos    Vec2    //gx:extern p
	Radius float64 //gx:extern r
}

//gx:extern c2Capsule
type Capsule struct {
	A, B   Vec2
	Radius float64 //gx:extern r
}

//gx:extern c2Poly
type Polygon struct {
	Count   int
	Verts   [8]Vec2
	Normals [8]Vec2 //gx:extern norms
}

//gx:extern c2MakePoly
func (p *Polygon) ConvexHull()

//gx:extern c2Norms
func c2Norms(verts *Vec2, norms *Vec2, count int)

func (p *Polygon) CalculateNormals() {
	c2Norms(&p.Verts[0], &p.Normals[0], p.Count)
}

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

//gx:extern c2CircletoPolyManifold
func c2CircletoPolyManifold(a Circle, b *Polygon, bx *c2x, m *IntersectResult)

func IntersectCirclePolygon(a Circle, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	c2CircletoPolyManifold(a, b, &bx, &result)
	return result
}

//gx:extern c2CapsuletoPolyManifold
func c2CapsuletoPolyManifold(a Capsule, b *Polygon, bx *c2x, m *IntersectResult)

func IntersectCapsulePolygon(a Capsule, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	c2CapsuletoPolyManifold(a, b, &bx, &result)
	return result
}

//gx:extern c2PolytoPolyManifold
func c2PolytoPolyManifold(a *Polygon, ax *c2x, b *Polygon, bx *c2x, m *IntersectResult)

func IntersectPolygons(a *Polygon, aPos Vec2, aAng float64, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	ax := c2x{aPos, c2r{Cos(aAng), Sin(aAng)}}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
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
