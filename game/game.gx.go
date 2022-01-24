package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

var gameCameraSize = Vec2{36, 20.25}.Scale(1.2)

var gameCamera = rl.Camera2D{
	Target: Vec2{0, 0},
}

var gameTime = 0.0
var deltaTime = 0.0

//
// Settings
//

var frictionDecel = 25.0

var playerSize = Vec2{1.5, 1}
var playerJumpStrength = 14.0
var playerGravityStrength = 28.0
var playerHorizontalControlsAccel = 17.0

var playerMinimumHorizontalSpeedForFriction = 12.0

const planetGravRadiusMultiplier = 1.38

//
// Init
//

func initGame() {
	if !edit.LoadSession() {
		// Test planet and player

		planetLay := Vec2{0, 24}
		planetRadius := 64.0
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
		lay.Rot = Atan2(up.Up.X, -up.Up.Y)
	})

	// Horizontal controls
	Each(func(ent Entity, player *Player, up *Up, vel *Velocity) {
		appliedControls := false
		if rl.IsKeyDown(rl.KEY_A) || rl.IsKeyDown(rl.KEY_LEFT) {
			dir := Vec2{up.Up.Y, -up.Up.X}
			vel.Vel = vel.Vel.Add(dir.Scale(playerHorizontalControlsAccel * deltaTime))
			appliedControls = true
		}
		if rl.IsKeyDown(rl.KEY_D) || rl.IsKeyDown(rl.KEY_RIGHT) {
			dir := Vec2{-up.Up.Y, up.Up.X}
			vel.Vel = vel.Vel.Add(dir.Scale(playerHorizontalControlsAccel * deltaTime))
			appliedControls = true
		}
		if appliedControls {
			upVel := up.Up.Scale(vel.Vel.DotProduct(up.Up))
			tangentVel := vel.Vel.Subtract(upVel)
			if tangentSpeed := tangentVel.Length(); tangentSpeed <= playerMinimumHorizontalSpeedForFriction {
				AddComponent(ent, DisableFriction{})
			}
		}
	})

	// Jump controls
	Each(func(ent Entity, player *Player, up *Up, vel *Velocity) {
		if rl.IsKeyPressed(rl.KEY_W) || rl.IsKeyPressed(rl.KEY_UP) {
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

				vel.Vel = vel.Vel.Add(dir.Scale(grav.Strength * deltaTime))
			}
		})
	})

	// Apply velocity
	Each(func(ent Entity, vel *Velocity, lay *Layout) {
		// Apply velocity
		lay.Pos = lay.Pos.Add(vel.Vel.Scale(deltaTime))

		// Handle collisions with planets
		Each(func(planetEnt Entity, planet *Planet, planetLay *Layout) {
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

				// Remove component of velocity along normal
				vel.Vel = vel.Vel.Subtract(in.Normal.Scale(vel.Vel.DotProduct(in.Normal)))

				// Mark for friction application
				AddComponent(ent, ApplyFriction{})
			}
		})
	})

	// Apply friction
	Each(func(ent Entity, applyFric *ApplyFriction, up *Up, vel *Velocity) {
		if !HasComponent[DisableFriction](ent) {
			upVel := up.Up.Scale(vel.Vel.DotProduct(up.Up))
			tangentVel := vel.Vel.Subtract(upVel)
			if tangentSpeed := tangentVel.Length(); tangentSpeed > 0 {
				tangentDir := tangentVel.Scale(1 / tangentSpeed)
				tangentSpeed = Max(0, tangentSpeed-frictionDecel*deltaTime)
				vel.Vel = upVel.Add(tangentDir.Scale(tangentSpeed))
			}
		}
	})
	ClearComponent[DisableFriction]()
	ClearComponent[ApplyFriction]()

	// Update camera
	Each(func(ent Entity, player *Player, lay *Layout, vel *Velocity) {
		if !player.CameraInitialized {
			player.CameraPos = lay.Pos
			player.CameraRot = lay.Rot
			player.CameraInitialized = true
		} else {
			// Position
			{
				lookAtDelta := vel.Vel.Scale(0.2)
				lookAt := lay.Pos.Add(lookAtDelta)
				rate := 14.0
				smoothing := 1 - Pow(2, -rate*deltaTime)
				player.CameraPos = player.CameraPos.Lerp(lookAt, smoothing)
			}

			// Rotation
			{
				targetRot := lay.Rot
				currRot := player.CameraRot
				if targetRotAbove := targetRot + 2*Pi; Abs(targetRotAbove-currRot) < Abs(targetRot-currRot) {
					targetRot = targetRotAbove
				}
				if targetRotBelow := targetRot - 2*Pi; Abs(targetRotBelow-currRot) < Abs(targetRot-currRot) {
					targetRot = targetRotBelow
				}
				rate := 5.0
				smoothing := 1 - Pow(2, -rate*deltaTime)
				player.CameraRot = player.CameraRot + smoothing*(targetRot-currRot)
			}
			player.CameraInitialized = true
		}
		gameCamera.Target = player.CameraPos
		if player.CameraRot > 2*Pi {
			player.CameraRot -= 2 * Pi
		}
		if player.CameraRot < 0 {
			player.CameraRot += 2 * Pi
		}
		gameCamera.Rotation = -player.CameraRot * 180 / Pi
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
		rl.DrawCircleSectorLines(lay.Pos, planet.Radius+1, 0, 360, 28, rl.White)
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
