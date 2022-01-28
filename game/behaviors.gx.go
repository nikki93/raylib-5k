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

	Strength float64 `default:"28"`
}

type CollisionShape struct {
	Behavior

	Verts []Vec2
}

type Player struct {
	Behavior

	JumpsRemaining int `default:"1"`
	lastJumpTime   float64

	SmoothedVel       Vec2
	CameraInitialized bool
	CameraPos         Vec2
	CameraRot         float64

	FlipH bool

	BeamOn             bool
	BeamEnd            Vec2
	BeamTimeSinceStart float64
	BeamTimeTillDamage float64

	ElementAmounts [NumElementTypes]int

	BuildUIEnabled        bool
	BuildUIPos            Vec2
	BuildUIMouseOver      bool
	BuildUISelectedTypeId ResourceTypeId `default:"-1"`
}

//
// Resource
//

type ElementTypeId int

const (
	CarbonElement     ElementTypeId = 0
	SiliconElement    ElementTypeId = 1
	FuelElement       ElementTypeId = 2
	AntimatterElement ElementTypeId = 3
	NumElementTypes                 = 4
)

type ElementType struct {
	Name          string
	IconImageName string
	iconTexture   rl.Texture
}

var elementTypes = func() [NumElementTypes]ElementType {
	result := [NumElementTypes]ElementType{}
	result[CarbonElement] = ElementType{
		Name:          "carbon",
		IconImageName: "icon_element_carbon.png",
	}
	result[SiliconElement] = ElementType{
		Name:          "silicon",
		IconImageName: "icon_element_silicon.png",
	}
	result[FuelElement] = ElementType{
		Name:          "fuel",
		IconImageName: "icon_element_fuel.png",
	}
	result[AntimatterElement] = ElementType{
		Name:          "antimatter",
		IconImageName: "icon_element_antimatter.png",
	}
	return result
}()

type ElementAmount struct {
	TypeId ElementTypeId
	Amount int
}

type ResourceTypeId int

type ResourceType struct {
	Name                   string
	ImageName              string
	NumFrames              int `default:"1"`
	IconImageName          string
	VerticalOffset         float64
	VerticalOffsetVariance float64

	Health         int `default:"3"`
	ElementAmounts []ElementAmount

	Buildable bool

	CollisionShapeVerts []Vec2

	texture     rl.Texture
	iconTexture rl.Texture
}

var resourceTypes = [...]ResourceType{
	{
		Name:                   "fungus_giant",
		ImageName:              "resource_fungus_giant.png",
		VerticalOffset:         -0.8,
		VerticalOffsetVariance: -0.2,
		Health:                 15,
		ElementAmounts: []ElementAmount{
			{TypeId: CarbonElement, Amount: 11},
		},
	},
	{
		Name:                   "fungus_tiny",
		ImageName:              "resource_fungus_tiny.png",
		VerticalOffset:         -0.3,
		VerticalOffsetVariance: -0.08,
		Health:                 3,
		ElementAmounts: []ElementAmount{
			{TypeId: CarbonElement, Amount: 2},
		},
	},
	{
		Name:                   "sprout_tiny",
		ImageName:              "resource_sprout_tiny.png",
		VerticalOffset:         -0.3,
		VerticalOffsetVariance: -0.08,
		Health:                 3,
		ElementAmounts: []ElementAmount{
			{TypeId: CarbonElement, Amount: 1},
		},
	},

	{
		Name:                   "rock_large",
		ImageName:              "resource_rock_large.png",
		VerticalOffset:         -0.8,
		VerticalOffsetVariance: -0.2,
		Health:                 50,
		ElementAmounts: []ElementAmount{
			{TypeId: SiliconElement, Amount: 18},
		},
	},
	{
		Name:                   "rock_medium",
		ImageName:              "resource_rock_medium.png",
		VerticalOffset:         -0.4,
		VerticalOffsetVariance: -0.2,
		Health:                 20,
		ElementAmounts: []ElementAmount{
			{TypeId: SiliconElement, Amount: 6},
		},
	},

	{
		Name:          "building_refiner",
		ImageName:     "resource_building_refiner.png",
		NumFrames:     2,
		IconImageName: "icon_building_refiner.png",
		Health:        50,
		ElementAmounts: []ElementAmount{
			{TypeId: SiliconElement, Amount: 100},
		},
		Buildable: true,
		CollisionShapeVerts: []Vec2{
			{-0.5 * 3, -2},
			{0.5 * 3, -2},
			{0.5 * 3, -0.8},
			{-0.5 * 3, -0.8},
		},
	},
	{
		Name:          "building_launchpad",
		ImageName:     "resource_building_refiner.png",
		NumFrames:     2,
		IconImageName: "icon_building_refiner.png",
		Health:        50,
		ElementAmounts: []ElementAmount{
			{TypeId: CarbonElement, Amount: 128},
			{TypeId: SiliconElement, Amount: 249},
			{TypeId: AntimatterElement, Amount: 2},
		},
		Buildable: true,
		CollisionShapeVerts: []Vec2{
			{-0.5 * 3, -2},
			{0.5 * 3, -2},
			{0.5 * 3, -0.8},
			{-0.5 * 3, -0.8},
		},
	},
}

type Resource struct {
	Behavior

	TypeId ResourceTypeId

	Frame int
	FlipH bool

	Health int
}

type ResourceDamaged struct {
	rotDeviation   float64
	time           float64
	lastDamageTime float64
}

//
// Interactables
//

type InteractionHint struct {
	Interactable   bool
	Message        string
	VerticalOffset float64 `default:"-4"`
}

type Refiner struct {
	Behavior

	CarbonAmount           int
	FuelAmount             int
	TimeTillNextRefinement float64
}

type Launchpad struct {
	Behavior
}
