package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
	"github.com/nikki93/raylib-5k/core/str"
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

var surfaceFrictionDecel = 25.0
var atmosphereFrictionDecel = 18.0

var playerSize = Vec2{1.5, 1}
var playerJumpStrength = 14.0
var playerGravityStrength = 28.0
var playerHorizontalControlsAccel = 17.0

var playerMinimumHorizontalSpeedForFriction = 12.0

var planetGravRadiusMultiplier = 1.38

//
// Init
//

type PlanetNoiseBand struct {
	Frequency, Amplitude float64
}

func createPlanet(pos Vec2, radius float64) Entity {
	ent := CreateEntity(
		Layout{
			Pos: pos,
		},
		Planet{
			Radius: radius,
		},
	)
	planet := GetComponent[Planet](ent)
	if planet == nil {
		str.Print("planet generation failed?")
	}

	// Generation parameters
	thickness := 1.0
	resolution := 2 * Pi * radius / thickness
	noiseBands := [...]PlanetNoiseBand{
		{Frequency: 0.003, Amplitude: 9.0},
		{Frequency: 0.015, Amplitude: 1.0},
	}

	// Generate heights and polygons
	angleStep := 2 * Pi / resolution
	prevInner := Vec2{0, 0}
	prevOuter := Vec2{0, 0}
	for angle := 0.0; angle < 2*Pi; angle += angleStep {
		// Height
		height := radius
		for _, band := range noiseBands {
			height += band.Amplitude * Noise1(resolution*band.Frequency*angle)
		}
		planet.terrain.heights = append(planet.terrain.heights, height)

		// Polygon vertices -- stitch with previous
		vertexAtHeight := func(height float64) Vec2 {
			return Vec2{Cos(angle), Sin(angle)}.Scale(height)
		}
		outer := vertexAtHeight(height)
		inner := vertexAtHeight(height - thickness)
		if angle > 0 {
			poly := Polygon{}
			poly.Count = 4
			poly.Verts[0] = prevInner
			poly.Verts[1] = prevOuter
			poly.Verts[2] = outer
			poly.Verts[3] = inner
			poly.CalculateNormals()
			planet.terrain.polys = append(planet.terrain.polys, poly)
		}
		prevOuter = outer
		prevInner = inner
	}

	return ent
}

func initGame() {
	if !edit.LoadSession() {
		// Home planet
		homePlanetPos := Vec2{0, 24}
		homePlanetRadius := 64.0
		createPlanet(homePlanetPos, homePlanetRadius)

		// Player
		playerPos := Vec2{0, homePlanetPos.Y - homePlanetRadius - 3 - 0.5*playerSize.Y}
		CreateEntity(
			Layout{
				Pos: playerPos,
			},
			Velocity{},
			Up{},
			Gravity{
				Strength: playerGravityStrength,
			},
			Player{},
		)
		edit.Camera().Target = playerPos

		// Smaller planet
		mediumPlanetRadius := 0.4 * homePlanetRadius
		createPlanet(
			Vec2{0, homePlanetPos.Y - 1.3*homePlanetRadius - 1.3*mediumPlanetRadius},
			mediumPlanetRadius,
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
	Each(func(ent Entity, lay *Layout) {
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
			up := AddComponent(ent, Up{})
			up.Up = dir.Negate().Normalize()
		} else {
			RemoveComponent[Up](ent)
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
			player.FlipHorizontal = true
			appliedControls = true
		}
		if rl.IsKeyDown(rl.KEY_D) || rl.IsKeyDown(rl.KEY_RIGHT) {
			dir := Vec2{-up.Up.Y, up.Up.X}
			vel.Vel = vel.Vel.Add(dir.Scale(playerHorizontalControlsAccel * deltaTime))
			player.FlipHorizontal = false
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
				AddComponent(ent, ApplySurfaceFriction{})
			}
		})
	})

	// Apply friction
	Each(func(ent Entity, up *Up, vel *Velocity) {
		if !HasComponent[DisableFriction](ent) {
			upVel := up.Up.Scale(vel.Vel.DotProduct(up.Up))
			tangentVel := vel.Vel.Subtract(upVel)
			if tangentSpeed := tangentVel.Length(); tangentSpeed > 0 {
				tangentDir := tangentVel.Scale(1 / tangentSpeed)
				decel := 0.0
				if HasComponent[ApplySurfaceFriction](ent) {
					decel = surfaceFrictionDecel
				} else {
					decel = atmosphereFrictionDecel
				}
				tangentSpeed = Max(0, tangentSpeed-decel*deltaTime)
				vel.Vel = upVel.Add(tangentDir.Scale(tangentSpeed))
			}
		}
	})
	ClearComponent[DisableFriction]()
	ClearComponent[ApplySurfaceFriction]()

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

//gx:extern getAssetPath
func getAssetPath(assetName string) string

var playerTexture = rl.LoadTexture(getAssetPath("player.png"))

func drawGame() {
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	// Planets
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		// Base radius / sea level
		rl.DrawCircleSector(lay.Pos, planet.Radius, 0, 360, 128, rl.Color{0x7a, 0x36, 0x7b, 0xff})

		// Polygons
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rl.CheckRenderBatchLimit(2 * 4 * len(planet.terrain.polys))
		rl.Begin(rl.Lines)
		for _, poly := range planet.terrain.polys {
			drawLine := func(a, b Vec2) {
				rl.Color4ub(0xff, 0xff, 0xff, 0xff)
				rl.Vertex2f(a.X, a.Y)
				rl.Vertex2f(b.X, b.Y)
			}
			drawLine(poly.Verts[0], poly.Verts[1])
			drawLine(poly.Verts[1], poly.Verts[2])
			drawLine(poly.Verts[2], poly.Verts[3])
			drawLine(poly.Verts[3], poly.Verts[0])
		}
		rl.End()
		rl.PopMatrix()

		// Gravity radius
		gravRadius := planetGravRadiusMultiplier * planet.Radius
		rl.DrawCircleSectorLines(lay.Pos, gravRadius, 0, 360, 32, rl.Color{0x7a, 0x36, 0x7b, 0xff})
	})

	// Player
	Each(func(ent Entity, player *Player, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rl.Rotatef(lay.Rot*180/Pi, 0, 0, 1)

		//rl.DrawRectangleRec(rl.Rectangle{
		//  X:      -0.5 * playerSize.X,
		//  Y:      -0.5 * playerSize.Y,
		//  Width:  playerSize.X,
		//  Height: playerSize.Y,
		//}, rl.Color{0xbe, 0x77, 0x2b, 0xff})

		texWidth := float64(playerTexture.Width)
		texHeight := float64(playerTexture.Height)
		texSource := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  texWidth,
			Height: texHeight,
		}
		destHeight := texHeight * playerSize.X / texWidth
		texDest := rl.Rectangle{
			X:      -0.5 * playerSize.X,
			Y:      0.5*playerSize.Y - destHeight,
			Width:  playerSize.X,
			Height: destHeight,
		}
		if player.FlipHorizontal {
			texSource.Width = -texSource.Width
		}
		rl.DrawTexturePro(playerTexture, texSource, texDest, Vec2{0, 0}, 0, rl.White)

		rl.PopMatrix()
	})
}
