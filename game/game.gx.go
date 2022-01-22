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
			Position{
				Pos: Vec2{400, 680},
			},
			Planet{
				Radius: 400,
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
}

//
// Draw
//

func drawGame() {
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	Each(func(ent Entity, planet *Planet, pos *Position) {
		rl.DrawCircleV(pos.Pos, planet.Radius, rl.Color{0x7a, 0x36, 0x7b, 0xff})
	})
}
