//gx:include "core/core.hh"

package game

import (
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

//
// Physics
//

type Layout struct {
	Behavior

	Pos Vec2
	Rot float64
}

type Velocity struct {
	Behavior

	Vel Vec2
}

type DisableFriction struct{}
type ApplySurfaceFriction struct{}

const numPlanetBitFrames = 9

type PlanetBit struct {
	Frame        int
	Rot          float64
	Perturb      Vec2
	FlipH, FlipV bool
}

const numPlanetBitsPerSegment = 8

type PlanetBits [numPlanetBitsPerSegment]PlanetBit

type Planet struct {
	Behavior

	Radius float64

	Verts []Vec2
	Bits  []PlanetBits
}

type Up struct {
	Behavior

	Up            Vec2
	GroundNormals []Vec2
	Smooth        Vec2

	lastGroundTime float64
}

type Gravity struct {
	Behavior

	Strength float64
}

type Player struct {
	Behavior

	SmoothedVel       Vec2
	CameraInitialized bool
	CameraPos         Vec2
	CameraRot         float64

	FlipH bool

	BeamOn   bool
	BeamEnd  Vec2
	BeamTime float64
}

//
// Resource
//

type ElementType int

const (
	CarbonElement   ElementType = 0
	SiliconElement  ElementType = 1
	NumElementTypes             = 2
)

type ElementAmount struct {
	Type   ElementType
	Amount int
}

type ResourceTypeId int

type ResourceType struct {
	Name                   string
	ImageName              string
	VerticalOffset         float64
	VerticalOffsetVariance float64
	ElementAmounts         []ElementAmount
	Texture                rl.Texture
}

var resourceTypes = [...]ResourceType{
	{
		Name:                   "fungus_giant",
		ImageName:              "resource_fungus_giant.png",
		VerticalOffset:         -0.8,
		VerticalOffsetVariance: -0.2,
		ElementAmounts: []ElementAmount{
			{Type: CarbonElement, Amount: 10},
		},
	},
	{
		Name:                   "fungus_tiny",
		ImageName:              "resource_fungus_tiny.png",
		VerticalOffset:         -0.3,
		VerticalOffsetVariance: -0.08,
		ElementAmounts: []ElementAmount{
			{Type: CarbonElement, Amount: 4},
		},
	},
	{
		Name:                   "sprout_tiny",
		ImageName:              "resource_sprout_tiny.png",
		VerticalOffset:         -0.3,
		VerticalOffsetVariance: -0.08,
		ElementAmounts: []ElementAmount{
			{Type: CarbonElement, Amount: 3},
		},
	},
	{
		Name:                   "rock_large",
		ImageName:              "resource_rock_large.png",
		VerticalOffset:         -0.8,
		VerticalOffsetVariance: -0.2,
		ElementAmounts: []ElementAmount{
			{Type: SiliconElement, Amount: 10},
		},
	},
	{
		Name:                   "rock_medium",
		ImageName:              "resource_rock_medium.png",
		VerticalOffset:         -0.4,
		VerticalOffsetVariance: -0.2,
		ElementAmounts: []ElementAmount{
			{Type: SiliconElement, Amount: 4},
		},
	},
}

type Resource struct {
	Behavior

	TypeId ResourceTypeId

	FlipH bool
}
