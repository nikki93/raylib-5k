package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
	"github.com/nikki93/raylib-5k/core/str"
)

var gameCameraZoom = 0.7
var gameCameraSize = Vec2{36, 20.25}.Scale(gameCameraZoom)

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

var spriteScale = playerSize.X / float64(playerTexture.Width)

//
// Init
//

func unitRandom() float64 {
	return float64(rl.GetRandomValue(0, 100)) / 100.0
}

type NoiseBand struct {
	Frequency, Amplitude float64
}

var bitsTextureBasic = rl.LoadTexture(getAssetPath("planet_surface_bits_basic.png"))

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
	segmentLength := 1.5 * playerSize.X
	resolution := 2 * Pi * radius / segmentLength
	noiseBands := [...]NoiseBand{
		{Frequency: 0.003, Amplitude: 0.2 * radius},
		{Frequency: 0.015, Amplitude: 0.015 * radius},
		{Frequency: 0.060, Amplitude: 0.015 * radius},
	}

	// Generate heights and vertices
	angleStep := 2 * Pi / resolution
	for angle := 0.0; angle < 2*Pi; angle += angleStep {
		// Height
		height := radius
		for _, band := range noiseBands {
			height += band.Amplitude * Noise1(resolution*band.Frequency*angle)
		}

		// Vertex
		vert := Vec2{Cos(angle), Sin(angle)}.Scale(height)
		planet.Verts = append(planet.Verts, vert)

		// Bits
		planet.Bits = append(planet.Bits, PlanetBits{})
		for _, bit := range planet.Bits[len(planet.Bits)-1] {
			bit.Frame = rl.GetRandomValue(0, numPlanetBitFrames-1)
			bit.Rot = 2 * Pi * unitRandom()
			bit.Perturb = Vec2{float64(rl.GetRandomValue(-1, 1)), float64(rl.GetRandomValue(-1, 1))}
			bit.FlipH = unitRandom() < 0.5
			bit.FlipV = unitRandom() < 0.5
		}
	}

	return ent
}

type CreateResourcesParams struct {
	TypeId     ResourceTypeId
	Planet     Entity
	Resolution int `default:"16"`
	NoiseBands []NoiseBand
	Exponent   float64 `default:"6"`
	Thinning   float64 `default:"1"`
}

func createResources(params CreateResourcesParams) {
	planet := GetComponent[Planet](params.Planet)
	planetLay := GetComponent[Layout](params.Planet)

	resourceType := &resourceTypes[params.TypeId]

	nVerts := len(planet.Verts)
	for vertIndex, localVertPos := range planet.Verts {
		vertPos := planetLay.Pos.Add(localVertPos)
		nextVertIndex := (vertIndex + 1) % nVerts
		nextVertPos := planetLay.Pos.Add(planet.Verts[nextVertIndex])
		edgeDelta := nextVertPos.Subtract(vertPos)
		edgeDir := Vec2{edgeDelta.Y, -edgeDelta.X}.Normalize()

		for resI := 0; resI < params.Resolution; resI++ {
			resF := float64(resI) / float64(params.Resolution)
			pos := vertPos.Add(edgeDelta.Scale(resF))
			localPos := pos.Subtract(planetLay.Pos)
			angle := Atan2(localPos.X, -localPos.Y) / (2 * Pi)

			noise := 0.0
			for _, band := range params.NoiseBands {
				noise += band.Amplitude * Sin(band.Frequency*angle)
			}
			probability := 0.5 + 0.5*noise
			probability = Pow(probability, params.Exponent)
			probability *= params.Thinning * 0.2

			roll := unitRandom()

			if roll < probability {
				upDir := pos.Normalize()
				dir := upDir.Lerp(edgeDir, 0.8)
				rot := Atan2(dir.X, -dir.Y) + Pi/9*(unitRandom()-0.5)
				rotDir := Vec2{Cos(rot), Sin(rot)}

				verticalOffsetDelta := resourceType.VerticalOffsetVariance * unitRandom()
				verticalOffsetDir := Vec2{rotDir.Y, -rotDir.X}
				pos = pos.Add(verticalOffsetDir.Scale(resourceType.VerticalOffset + verticalOffsetDelta))

				CreateEntity(
					Layout{
						Pos: pos,
						Rot: rot,
					},
					Resource{
						TypeId: params.TypeId,
						FlipH:  unitRandom() < 0.5,
					},
				)
			}
		}
	}
}

func resourceTypeIdForName(name string) ResourceTypeId {
	for i, resourceType := range resourceTypes {
		if resourceType.Name == name {
			return ResourceTypeId(i)
		}
	}
	return ResourceTypeId(-1)
}

func initGame() {
	rl.SetRandomSeed(1024)

	// Initialize resource textures
	for _, resourceType := range resourceTypes {
		resourceType.Texture = rl.LoadTexture(getAssetPath(resourceType.ImageName))
	}

	// Scene
	if !edit.LoadSession() {
		// Home planet
		homePlanetPos := Vec2{0, 24}
		homePlanetRadius := 64.0
		homePlanet := createPlanet(homePlanetPos, homePlanetRadius)

		// Player
		playerPos := Vec2{0, homePlanetPos.Y - homePlanetRadius - 0.5*playerSize.Y - 5}
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
		mediumPlanetRadius := 0.6 * homePlanetRadius
		createPlanet(
			Vec2{0, homePlanetPos.Y - 1.3*homePlanetRadius - 1.3*mediumPlanetRadius},
			mediumPlanetRadius,
		)

		// Resources
		createResources(CreateResourcesParams{
			TypeId: resourceTypeIdForName("fungus_tiny"),
			Planet: homePlanet,
			NoiseBands: []NoiseBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Thinning: 0.6,
		})
		createResources(CreateResourcesParams{
			TypeId: resourceTypeIdForName("sprout_tiny"),
			Planet: homePlanet,
			NoiseBands: []NoiseBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Exponent: 2,
		})
		createResources(CreateResourcesParams{
			TypeId: resourceTypeIdForName("fungus_giant"),
			Planet: homePlanet,
			NoiseBands: []NoiseBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Exponent: 1,
			Thinning: 0.02,
		})

		edit.SaveSnapshot("initialize scene")
	}
}

//
// Update
//

func updateGame(dt float64) {
	gameTime += dt
	deltaTime = dt

	// Update up direction and clear ground normals
	Each(func(ent Entity, player *Player, lay *Layout) {
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
			up.GroundNormals = []Vec2{}
		} else {
			RemoveComponent[Up](ent)
		}
	})

	// Horizontal controls
	Each(func(ent Entity, player *Player, up *Up, vel *Velocity) {
		appliedControls := false
		if rl.IsKeyDown(rl.KEY_A) || rl.IsKeyDown(rl.KEY_LEFT) {
			dir := Vec2{up.Up.Y, -up.Up.X}
			vel.Vel = vel.Vel.Add(dir.Scale(playerHorizontalControlsAccel * deltaTime))
			player.FlipH = true
			appliedControls = true
		}
		if rl.IsKeyDown(rl.KEY_D) || rl.IsKeyDown(rl.KEY_RIGHT) {
			dir := Vec2{-up.Up.Y, up.Up.X}
			vel.Vel = vel.Vel.Add(dir.Scale(playerHorizontalControlsAccel * deltaTime))
			player.FlipH = false
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
		up := GetComponent[Up](ent)

		// Apply velocity
		lay.Pos = lay.Pos.Add(vel.Vel.Scale(deltaTime))

		// Handle collisons
		thickness := 0.4
		poly := Polygon{}
		poly.Count = 4
		reducedPlayerSize := playerSize.Subtract(Vec2{2 * thickness, 2.14 * thickness})
		poly.Verts[0] = Vec2{-0.5 * reducedPlayerSize.X, -0.5 * reducedPlayerSize.Y}
		poly.Verts[1] = Vec2{0.5 * reducedPlayerSize.X, -0.5 * reducedPlayerSize.Y}
		poly.Verts[2] = Vec2{0.5 * reducedPlayerSize.X, 0.5 * reducedPlayerSize.Y}
		poly.Verts[3] = Vec2{-0.5 * reducedPlayerSize.X, 0.5 * reducedPlayerSize.Y}
		poly.CalculateNormals()
		Each(func(planetEnt Entity, planet *Planet, planetLay *Layout) {
			nVerts := len(planet.Verts)
			for i, localVertPos := range planet.Verts {
				a := planetLay.Pos.Add(localVertPos)
				b := planetLay.Pos.Add(planet.Verts[(i+1)%nVerts])
				capsule := Capsule{A: a, B: b, Radius: thickness}

				// Calculate intersection
				in := IntersectCapsulePolygon(capsule, &poly, lay.Pos, lay.Rot)
				if in.Count > 0 {
					// Push position out by intersection
					lay.Pos = lay.Pos.Add(in.Normal.Scale(in.Depths[0]))

					// Remove component of velocity along normal
					vel.Vel = vel.Vel.Subtract(in.Normal.Scale(vel.Vel.DotProduct(in.Normal)))

					// Mark for friction application
					AddComponent(ent, ApplySurfaceFriction{})

					// Track normal
					if up != nil && in.Normal.DotProduct(up.Up) > 0.2 {
						up.GroundNormals = append(up.GroundNormals, in.Normal)
						up.lastGroundTime = gameTime
					}
				}
			}
		})
	})

	// Update smooth up direction
	Each(func(ent Entity, up *Up, lay *Layout) {
		target := Vec2{0, 0}
		rate := 28.0
		if len(up.GroundNormals) == 0 {
			if gameTime-up.lastGroundTime < 0.3 {
				rate = 7.0
			}
			target = up.Up
		} else {
			for _, groundNormal := range up.GroundNormals {
				target = target.Add(groundNormal)
			}
			target = target.Scale(1 / float64(len(up.GroundNormals)))
		}
		smoothing := 1 - Pow(2, -rate*deltaTime)
		up.Smooth = up.Smooth.Add(target.Subtract(up.Smooth).Scale(smoothing)).Normalize()
		lay.Rot = Atan2(up.Smooth.X, -up.Smooth.Y)
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
		{
			// Smoothed velocity
			rate := 160.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			player.SmoothedVel = player.SmoothedVel.Lerp(vel.Vel, smoothing)
		}
		upOffset := Vec2{0, 0}
		velOffset := player.SmoothedVel.Scale(0.4 * gameCameraZoom)
		if up := GetComponent[Up](ent); up != nil {
			upOffset = up.Up.Scale(0.7)
			upVelOffset := up.Up.Scale(velOffset.DotProduct(up.Up))
			velOffset = velOffset.Subtract(upVelOffset).Add(upVelOffset.Scale(0.1))
		}
		lookAt := lay.Pos.Add(upOffset).Add(velOffset)

		if !player.CameraInitialized {
			player.CameraPos = lookAt
			player.CameraRot = lay.Rot
			player.CameraInitialized = true
		} else {
			// Position
			{
				zoomDelta := 1 - gameCameraZoom
				rateFactor := zoomDelta * zoomDelta * zoomDelta
				rate := 14.0 / (1 - rateFactor)
				smoothing := 1 - Pow(2, -rate*deltaTime)
				player.CameraPos = player.CameraPos.Lerp(lookAt, smoothing)
			}

			// Rotation
			{
				targetRot := lay.Rot
				if up := GetComponent[Up](ent); up != nil {
					targetRot = Atan2(up.Up.X, -up.Up.Y)
				}
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

var whiteTexture = rl.LoadTextureFromImage(rl.GenImageColor(1, 1, rl.Color{0xff, 0xff, 0xff, 0xff}))

var playerTexture = rl.LoadTexture(getAssetPath("player.png"))

func drawGame() {
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	// Resources
	Each(func(ent Entity, resource *Resource, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rl.Rotatef(lay.Rot*180/Pi, 0, 0, 1)

		resourceType := &resourceTypes[resource.TypeId]
		texture := resourceType.Texture
		texSize := Vec2{float64(texture.Width), float64(texture.Height)}
		texSource := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  texSize.X,
			Height: texSize.Y,
		}
		destSize := texSize.Scale(spriteScale)
		texDest := rl.Rectangle{
			X:      -0.5 * destSize.X,
			Y:      -destSize.Y,
			Width:  destSize.X,
			Height: destSize.Y,
		}
		if resource.FlipH {
			texSource.Width = -texSource.Width
		}
		rl.DrawTexturePro(texture, texSource, texDest, Vec2{0, 0}, 0, rl.White)

		rl.PopMatrix()
	})

	// Planet terrain
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)

		nVerts := len(planet.Verts)

		rl.CheckRenderBatchLimit(4 * (nVerts - 1))
		rl.SetTexture(whiteTexture.Id)
		rl.Begin(rl.Quads)
		for i := 1; i < nVerts; i++ {
			drawLine := func(a, b Vec2) {
				rl.Color4ub(0x15, 0x1d, 0x28, 0xff)
				rl.TexCoord2f(0, 0)
				rl.Vertex2f(0, 0)
				rl.TexCoord2f(1, 0)
				rl.Vertex2f(0, 0)
				rl.TexCoord2f(1, 1)
				rl.Vertex2f(b.X, b.Y)
				rl.TexCoord2f(0, 1)
				rl.Vertex2f(a.X, a.Y)
			}
			drawLine(planet.Verts[i-1], planet.Verts[i])
		}
		rl.End()

		rl.PopMatrix()
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
		destHeight := texHeight * spriteScale
		texDest := rl.Rectangle{
			X:      -0.5 * playerSize.X,
			Y:      0.5*playerSize.Y - destHeight,
			Width:  playerSize.X,
			Height: destHeight,
		}
		if player.FlipH {
			texSource.Width = -texSource.Width
		}
		rl.DrawTexturePro(playerTexture, texSource, texDest, Vec2{0, 0}, 0, rl.White)

		rl.PopMatrix()
	})

	// Planet bits
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)

		nVerts := len(planet.Verts)

		nBits := 0

		bitTex := bitsTextureBasic
		bitTexHeight := float64(bitTex.Height)
		bitTexSource := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  bitTexHeight,
			Height: bitTexHeight,
		}
		bitTexDest := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  spriteScale * bitTexHeight,
			Height: spriteScale * bitTexHeight,
		}
		bitTexOrigin := Vec2{bitTexDest.Width, bitTexDest.Height}.Scale(0.5)
		for vertI, vertPos := range planet.Verts {
			edgeDelta := planet.Verts[(vertI+1)%nVerts].Subtract(vertPos)
			for bitI, bit := range planet.Bits[vertI] {
				bitTexSource.X = bitTexHeight * float64(bit.Frame)

				frac := float64(bitI) / float64(numPlanetBitsPerSegment)
				perturb := bit.Perturb.Scale(spriteScale / bitTexDest.Width)
				pos := vertPos.Add(edgeDelta.Scale(frac)).Add(perturb)

				bitTexDest.X = pos.X
				bitTexDest.Y = pos.Y

				rl.DrawCircleV(pos, 0.1, rl.Color{})

				rl.DrawTexturePro(bitTex, bitTexSource, bitTexDest, bitTexOrigin, bit.Rot, rl.Color{0x4d, 0x2b, 0x32, 0xff})

				nBits++
			}
		}

		str.Display("%d bits", nBits)

		rl.PopMatrix()
	})

}
