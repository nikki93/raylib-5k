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

//gx:extern c2Ray
type Ray struct {
	Position            Vec2    //gx:extern p
	NormalizedDirection Vec2    //gx:extern d
	Length              float64 //gx:extern t
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

//gx:extern c2CapsuletoCapsuleManifold
func c2CapsuletoCapsuleManifold(a Capsule, b Capsule, result *IntersectResult)

func IntersectCapsules(a Capsule, b Capsule) IntersectResult {
	result := IntersectResult{}
	c2CapsuletoCapsuleManifold(a, b, &result)
	return result
}

//gx:extern c2CircletoPolyManifold
func c2CircletoPolyManifold(a Circle, b *Polygon, bx *c2x, result *IntersectResult)

func IntersectCirclePolygon(a Circle, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	c2CircletoPolyManifold(a, b, &bx, &result)
	return result
}

//gx:extern c2CapsuletoPolyManifold
func c2CapsuletoPolyManifold(a Capsule, b *Polygon, bx *c2x, result *IntersectResult)

func IntersectCapsulePolygon(a Capsule, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	c2CapsuletoPolyManifold(a, b, &bx, &result)
	return result
}

//gx:extern c2PolytoPolyManifold
func c2PolytoPolyManifold(a *Polygon, ax *c2x, b *Polygon, bx *c2x, result *IntersectResult)

func IntersectPolygons(a *Polygon, aPos Vec2, aAng float64, b *Polygon, bPos Vec2, bAng float64) IntersectResult {
	result := IntersectResult{}
	ax := c2x{aPos, c2r{Cos(aAng), Sin(aAng)}}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	c2PolytoPolyManifold(a, &ax, b, &bx, &result)
	return result
}

//gx:extern c2Raycast
type c2Raycast struct {
	Distance float64 //gx:extern t
	Normal   Vec2    //gx:extern n
}

type RaycastResult struct {
	Hit      bool
	Distance float64
	Normal   Vec2
}

//gx:extern c2RaytoCapsule
func c2RaytoCapsule(a Ray, b Capsule, raycast *c2Raycast) int

func RaycastCapsule(ray Ray, capsule Capsule) RaycastResult {
	raycast := c2Raycast{}
	if c2RaytoCapsule(ray, capsule, &raycast) == 0 {
		return RaycastResult{Hit: false, Distance: 0, Normal: Vec2{0, 0}}
	}
	return RaycastResult{Hit: true, Distance: raycast.Distance, Normal: raycast.Normal}
}

//gx:extern c2RaytoPoly
func c2RaytoPoly(a Ray, b *Polygon, bx *c2x, raycast *c2Raycast) int

func RaycastPolygon(ray Ray, b *Polygon, bPos Vec2, bAng float64) RaycastResult {
	raycast := c2Raycast{}
	bx := c2x{bPos, c2r{Cos(bAng), Sin(bAng)}}
	if c2RaytoPoly(ray, b, &bx, &raycast) == 0 {
		return RaycastResult{Hit: false, Distance: 0, Normal: Vec2{0, 0}}
	}
	return RaycastResult{Hit: true, Distance: raycast.Distance, Normal: raycast.Normal}
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
