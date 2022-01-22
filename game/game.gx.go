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
// Settings
//

var playerSize = Vec2{24, 32}

//
// Init
//

func initGame() {
	if !edit.LoadSession() {
		// Test planet
		CreateEntity(
			Position{
				Pos: Vec2{400, 700},
			},
			Planet{
				Radius: 400,
			},
		)

		// Player
		CreateEntity(
			Position{
				Pos: Vec2{400, 300 - 0.5*playerSize.Y},
			},
			Player{},
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

	// Planets
	Each(func(ent Entity, planet *Planet, pos *Position) {
		rl.DrawCircleV(pos.Pos, planet.Radius, rl.Color{0x7a, 0x36, 0x7b, 0xff})
	})

	// Player
	Each(func(ent Entity, player *Player, pos *Position) {
		rl.DrawRectangleRec(rl.Rectangle{
			X:      pos.Pos.X - 0.5*playerSize.X,
			Y:      pos.Pos.Y - 0.5*playerSize.Y,
			Width:  playerSize.X,
			Height: playerSize.Y,
		}, rl.Color{0xbe, 0x77, 0x2b, 0xff})
	})
}
