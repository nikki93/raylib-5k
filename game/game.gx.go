//gx:include "core/core.hh"

package game

import (
	. "github.com/nikki93/dream-hotel/core/entity"
	. "github.com/nikki93/dream-hotel/core/geom"
	"github.com/nikki93/dream-hotel/core/rl"
)

var gameCameraSize = Vec2{864, 486}
var gameCamera = rl.Camera2D{
	Target: gameCameraSize.Scale(0.5),
}

var gameTime = 0.0
var deltaTime = 0.0

type Circle struct {
	Behavior

	Pos    Vec2
	Radius float64
}

//
// Init
//

func initGame() {
	CreateEntity(
		Circle{
			Pos:    Vec2{100, 200},
			Radius: 20,
		},
	)
}

//
// Update
//

func updateGame(dt float64) {
	gameTime += dt
	deltaTime = dt

	Each(func(ent Entity, circle *Circle) {
		circle.Pos.Y += 80 * Sin(3*rl.GetTime()) * dt
	})
}

//
// Draw
//

func drawGame() {
	Each(func(ent Entity, circle *Circle) {
		rl.DrawCircleV(circle.Pos, circle.Radius, rl.Red)
	})

	rl.DrawRectangleLines(0, 0, 864, 486, rl.Red)
}
