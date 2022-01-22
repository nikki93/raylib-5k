package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

var gameCameraSize = Vec2{864, 486}
var gameCamera = rl.Camera2D{
	Target: gameCameraSize.Scale(0.5),
}

var gameTime = 0.0
var deltaTime = 0.0

//
// Init
//

func initGame() {
	if !edit.LoadSession() {
		CreateEntity(
			Circle{
				Pos:    Vec2{100, 200},
				Radius: 20,
			},
		)
		edit.SaveSnapshot("initialize scene")
	}
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
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	Each(func(ent Entity, circle *Circle) {
		rl.DrawCircleV(circle.Pos, circle.Radius, rl.Red)
	})
}
