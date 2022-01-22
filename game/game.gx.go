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
var playerJumpStrength = 6.0
var playerGravityStrength = 9.0

const planetGravRadiusMultiplier = 1.38

//
// Init
//

func initGame() {
	if !edit.LoadSession() {
		// Test planet and player

		planetLay := Vec2{0, 24}
		planetRadius := 21.0
		CreateEntity(
			Layout{
				Pos: planetLay,
			},
			Planet{
				Radius: planetRadius,
			},
		)

		CreateEntity(
			Layout{
				Pos: Vec2{0, planetLay.Y - planetRadius - 3 - 0.5*playerSize.Y},
			},
			Velocity{},
			Up{},
			Gravity{
				Strength: playerGravityStrength,
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

	// Update up direction
	Each(func(ent Entity, up *Up, lay *Layout) {
		minSqDist := -1.0
		minDelta := Vec2{0, 0}
		Each(func(ent Entity, planet *Planet, planetLay *Layout) {
			delta := planetLay.Pos.Subtract(lay.Pos)
			sqDist := delta.LengthSqr()
			if minSqDist < 0 || sqDist < minSqDist {
				gravRadius := planetGravRadiusMultiplier * planet.Radius
				if sqDist < gravRadius*gravRadius {
					minSqDist = sqDist
					minDelta = delta
				}
			}
		})
		if minSqDist > 0 {
			dir := minDelta.Scale(1 / Sqrt(minSqDist))
			up.Up = dir.Negate().Normalize()
		}
	})

	// Rotate toward up direction
	Each(func(ent Entity, up *Up, lay *Layout) {
		lay.Rot = Atan2(up.Up.Y, up.Up.X)
	})

	// Jumping
	Each(func(ent Entity, player *Player, up *Up, vel *Velocity) {
		if rl.IsKeyPressed(rl.KEY_W) {
			vel.Vel = vel.Vel.Add(up.Up.Scale(playerJumpStrength))
		}
	})

	// Gravity toward planets
	Each(func(ent Entity, grav *Gravity, vel *Velocity, lay *Layout) {
		// Add to velocity
		Each(func(ent Entity, planet *Planet, planetLay *Layout) {
			delta := planetLay.Pos.Subtract(lay.Pos)
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
	Each(func(ent Entity, vel *Velocity, lay *Layout) {
		// Apply velocity
		lay.Pos = lay.Pos.Add(vel.Vel.Scale(dt))

		// Handle collisions with planets
		Each(func(ent Entity, planet *Planet, planetLay *Layout) {
			// Calculate intersection
			circ := Circle{Pos: planetLay.Pos, Radius: planet.Radius}
			poly := Polygon{}
			poly.Count = 4
			poly.Verts[0] = Vec2{-0.5 * playerSize.X, -0.5 * playerSize.Y}
			poly.Verts[1] = Vec2{0.5 * playerSize.X, -0.5 * playerSize.Y}
			poly.Verts[2] = Vec2{0.5 * playerSize.X, 0.5 * playerSize.Y}
			poly.Verts[3] = Vec2{-0.5 * playerSize.X, 0.5 * playerSize.Y}
			poly.CalculateNormals()
			in := IntersectCirclePolygon(circ, &poly, lay.Pos, lay.Rot)
			if in.Count > 0 {
				// Push position out by intersection
				lay.Pos = lay.Pos.Add(in.Normal.Scale(in.Depths[0]))

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
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		rl.DrawCircleSector(lay.Pos, planet.Radius, 0, 360, 128, rl.Color{0x7a, 0x36, 0x7b, 0xff})
		gravRadius := planetGravRadiusMultiplier * planet.Radius
		rl.DrawCircleSectorLines(lay.Pos, gravRadius, 0, 360, 32, rl.Color{0x7a, 0x36, 0x7b, 0xff})
	})

	// Player
	Each(func(ent Entity, player *Player, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rl.Rotatef(lay.Rot*180/Pi, 0, 0, 1)
		rl.DrawRectangleRec(rl.Rectangle{
			X:      -0.5 * playerSize.X,
			Y:      -0.5 * playerSize.Y,
			Width:  playerSize.X,
			Height: playerSize.Y,
		}, rl.Color{0xbe, 0x77, 0x2b, 0xff})
		rl.PopMatrix()
	})
}
