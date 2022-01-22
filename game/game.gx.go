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
			Velocity{},
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

	// Up direction
	Each(func(ent Entity, up *Up, pos *Position) {
		Each(func(ent Entity, planet *Planet, planetPos *Position) {
		})
	})

	// Jumping
	Each(func(ent Entity, player *Player, vel *Velocity) {
		if rl.IsKeyPressed(rl.KEY_W) {
		}
	})

	// Gravity toward planets
	Each(func(ent Entity, grav *Gravity, vel *Velocity, pos *Position) {
		// Add to velocity
		Each(func(ent Entity, planet *Planet, planetPos *Position) {
			delta := planetPos.Pos.Subtract(pos.Pos)
			sqDist := delta.LengthSqr()
			gravRadius := planetGravRadiusMultiplier * planet.Radius
			if sqDist < gravRadius*gravRadius {
				dist := Sqrt(sqDist)
				dir := delta.Scale(1 / dist)

				vel.Vel = vel.Vel.Add(dir.Scale(grav.Strength * dt))
			}
		})
	})

	// Apply velocity
	Each(func(ent Entity, vel *Velocity, pos *Position) {
		// Apply velocity
		pos.Pos = pos.Pos.Add(vel.Vel.Scale(dt))

		// Handle collisions with planets
		Each(func(ent Entity, planet *Planet, planetPos *Position) {
			// Calculate intersection
			circ := Circle{Pos: planetPos.Pos, Radius: planet.Radius}
			poly := Polygon{}
			poly.Count = 4
			poly.Verts[0] = Vec2{-0.5 * playerSize.X, -0.5 * playerSize.Y}
			poly.Verts[1] = Vec2{0.5 * playerSize.X, -0.5 * playerSize.Y}
			poly.Verts[2] = Vec2{0.5 * playerSize.X, 0.5 * playerSize.Y}
			poly.Verts[3] = Vec2{-0.5 * playerSize.X, 0.5 * playerSize.Y}
			poly.CalculateNormals()
			in := IntersectCirclePolygon(circ, &poly, pos.Pos)
			if in.Count > 0 {
				// Push position out by intersection
				pos.Pos = pos.Pos.Add(in.Normal.Scale(in.Depths[0]))

				// Remove component of vector along normal
				vel.Vel = vel.Vel.Subtract(in.Normal.Scale(vel.Vel.DotProduct(in.Normal)))
			}
		})
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
