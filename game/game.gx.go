package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

var gameCameraSize = Vec2{36, 20.25}

var gameCamera = rl.Camera2D{
	Target: Vec2{0, 0},
}

var gameTime = 0.0
var deltaTime = 0.0

//
// Settings
//

var playerSize = Vec2{1, 1}

const planetGravRadiusMultiplier = 1.38

//
// Init
//

func initGame() {
	if !edit.LoadSession() {
		// Test planet and player

		planetPos := Vec2{0, 24}
		planetRadius := 21.0
		CreateEntity(
			Position{
				Pos: planetPos,
			},
			Planet{
				Radius: planetRadius,
			},
		)

		CreateEntity(
			Position{
				Pos: Vec2{0, planetPos.Y - planetRadius - 3 - 0.5*playerSize.Y},
			},
			Gravity{
				Strength: 9,
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

	// Gravity toward planets
	Each(func(ent Entity, grav *Gravity, pos *Position) {
		Each(func(ent Entity, planet *Planet, planetPos *Position) {
			delta := planetPos.Pos.Subtract(pos.Pos)
			sqDist := delta.LengthSqr()
			gravRadius := planetGravRadiusMultiplier * planet.Radius
			if sqDist < gravRadius*gravRadius {
				dist := Sqrt(sqDist)
				dir := delta.Scale(1 / dist)

				grav.vel = grav.vel.Add(dir.Scale(grav.Strength * dt))
			}
		})

		pos.Pos = pos.Pos.Add(grav.vel.Scale(dt))
	})
}

//
// Draw
//

func drawGame() {
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	// Planets
	Each(func(ent Entity, planet *Planet, pos *Position) {
		rl.DrawCircleSector(pos.Pos, planet.Radius, 0, 360, 128, rl.Color{0x7a, 0x36, 0x7b, 0xff})
		gravRadius := planetGravRadiusMultiplier * planet.Radius
		rl.DrawCircleSectorLines(pos.Pos, gravRadius, 0, 360, 32, rl.Color{0x7a, 0x36, 0x7b, 0xff})
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
