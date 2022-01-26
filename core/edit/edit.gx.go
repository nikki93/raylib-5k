package edit

import (
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

//
// State
//

//gx:extern edit
var edit struct {
	enabled bool

	camera rl.Camera2D

	lineThickness float64

	mouseScreenPos  Vec2
	mouseWorldPos   Vec2
	mouseWorldDelta Vec2
}

func Enabled() bool {
	return edit.enabled
}

//gx:extern isEditMode
func IsMode(mode string) bool

//gx:extern setEditMode
func SetMode(mode string)

func Camera() *rl.Camera2D {
	return &edit.camera
}

func LineThickness() float64 {
	return edit.lineThickness
}

func MouseScreenPos() Vec2 {
	return edit.mouseScreenPos
}

func MouseWorldPos() Vec2 {
	return edit.mouseWorldPos
}

func MouseWorldDelta() Vec2 {
	return edit.mouseWorldDelta
}

//
// Selection
//

//gx:extern EditSelect
type Select struct{}

//
// Boxes
//

//gx:extern EditBox
type Box struct {
	Rect rl.Rectangle
}

//gx:extern mergeEditBox
func MergeBox(ent Entity, rect rl.Rectangle)

//
// Moves
//

//gx:extern EditMove
type Move struct {
	Delta Vec2
}

//
// Inspect
//

//gx:extern EditInspectContext
type InspectContext struct {
	ent     Entity
	changed bool
}

//
// Session
//

//gx:extern saveEditSnapshot
func SaveSnapshot(desc string)

//gx:extern loadEditSession
func LoadSession() bool

//gx:extern openSceneEdit
func OpenScene(assetName string)
