package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
	"github.com/nikki93/raylib-5k/core/str"
)

var gameCameraZoomTarget = 0.7
var gameCameraZoom = 0.7
var gameCameraBaseSize = Vec2{36, 20.25}
var gameCameraSize = gameCameraBaseSize.Scale(gameCameraZoom)

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
var playerHorizontalControlsAccel = 17.0
var playerMinimumHorizontalSpeedForFriction = 12.0
var playerJumpCooldown = 0.1

var planetGravRadiusMultiplier = 1.38
var planetSegmentThickness = 0.4

var spriteScale = playerSize.X / float64(playerTexture.Width)

var beamZapTime = 0.06
var beamDamagePeriod = 0.2
var beamDamage = 1

var refinerCarbonPerFuel = 5
var refinerRefinmentPeriod = 5.0

//
// Sounds
//

var music = rl.LoadMusicStream(getAssetPath("music_1.ogg"))

var laserSound = rl.LoadMusicStream(getAssetPath("sfx_laser_on.ogg")) // Music so it loops

var hitSound1 = rl.LoadSound(getAssetPath("sfx_hit_1.wav"))
var hitSound2 = rl.LoadSound(getAssetPath("sfx_hit_2.wav"))

var resourceHitGroundSound = rl.LoadSound(getAssetPath("sfx_vehicle_collision.ogg"))

//
// Init
//

func unitRandom() float64 {
	return float64(rl.GetRandomValue(0, 100)) / 100.0
}

type FrequencyBand struct {
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
	segmentLength := 1.5 * playerSize.X
	resolution := 2 * Pi * radius / segmentLength
	frequencyBands := [...]FrequencyBand{
		{Frequency: 0.003, Amplitude: 0.2 * radius},
		{Frequency: 0.015, Amplitude: 0.015 * radius},
		{Frequency: 0.060, Amplitude: 0.015 * radius},
	}

	// Generate heights and vertices
	angleStep := 2 * Pi / resolution
	for angle := 0.0; angle < 2*Pi; angle += angleStep {
		// Height
		height := radius
		for _, band := range frequencyBands {
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
	TypeName       string
	TypeId         ResourceTypeId
	Planet         Entity
	Resolution     int `default:"16"`
	FrequencyBands []FrequencyBand
	Exponent       float64 `default:"6"`
	Thinning       float64 `default:"1"`
}

func createResources(params CreateResourcesParams) {
	if params.TypeName != "" {
		params.TypeId = resourceTypeIdForName(params.TypeName)
	}

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
			for _, band := range params.FrequencyBands {
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
						Health: resourceType.Health,
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
	rl.HideCursor()

	// Initialize random seed
	rl.SetRandomSeed(1024)

	// Initialize element and resource textures
	for _, elementType := range elementTypes {
		elementType.iconTexture = rl.LoadTexture(getAssetPath(elementType.IconImageName))
	}
	for _, resourceType := range resourceTypes {
		resourceType.texture = rl.LoadTexture(getAssetPath(resourceType.ImageName))
		if resourceType.IconImageName != "" {
			resourceType.iconTexture = rl.LoadTexture(getAssetPath(resourceType.IconImageName))
		}
	}

	// Scene
	if !edit.LoadSession() {
		// Home planet
		homePlanetPos := Vec2{0, 24}
		homePlanetRadius := 64.0
		homePlanet := createPlanet(homePlanetPos, homePlanetRadius)

		// Player
		playerPos := Vec2{0, homePlanetPos.Y - homePlanetRadius - 0.5*playerSize.Y - 5}
		playerPolySize := playerSize.Subtract(Vec2{2 * planetSegmentThickness, 2.14 * planetSegmentThickness})
		CreateEntity(
			Layout{
				Pos: playerPos,
			},
			Velocity{},
			Up{},
			Gravity{},
			CollisionShape{
				Verts: []Vec2{
					{-0.5 * playerPolySize.X, -0.5 * playerPolySize.Y},
					{0.5 * playerPolySize.X, -0.5 * playerPolySize.Y},
					{0.5 * playerPolySize.X, 0.5 * playerPolySize.Y},
					{-0.5 * playerPolySize.X, 0.5 * playerPolySize.Y},
				},
			},
			Player{
				ElementAmounts: [NumElementTypes]int{100, 200},
			},
		)
		edit.Camera().Target = playerPos

		// Smaller planet
		mediumPlanetRadius := 0.3 * homePlanetRadius
		createPlanet(
			Vec2{0, homePlanetPos.Y - 1.2*homePlanetRadius - 1.2*mediumPlanetRadius},
			mediumPlanetRadius,
		)

		// Resources on home planet
		createResources(CreateResourcesParams{
			TypeName: "fungus_tiny",
			Planet:   homePlanet,
			FrequencyBands: []FrequencyBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Thinning: 0.6,
		})
		createResources(CreateResourcesParams{
			TypeName: "sprout_tiny",
			Planet:   homePlanet,
			FrequencyBands: []FrequencyBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Exponent: 2,
		})
		createResources(CreateResourcesParams{
			TypeName: "fungus_giant",
			Planet:   homePlanet,
			FrequencyBands: []FrequencyBand{
				{Frequency: 80, Amplitude: 0.5},
				{Frequency: 16, Amplitude: 0.5},
			},
			Exponent: 1,
			Thinning: 0.02,
		})
		createResources(CreateResourcesParams{
			TypeName: "rock_large",
			Planet:   homePlanet,
			FrequencyBands: []FrequencyBand{
				{Frequency: 60, Amplitude: 0.5},
				{Frequency: 3, Amplitude: 0.2},
			},
			Exponent: 1,
			Thinning: 0.001,
		})
		createResources(CreateResourcesParams{
			TypeName: "rock_medium",
			Planet:   homePlanet,
			FrequencyBands: []FrequencyBand{
				{Frequency: 60, Amplitude: 0.5},
				{Frequency: 3, Amplitude: 0.4},
			},
			Exponent: 1,
			Thinning: 0.015,
		})

		edit.SaveSnapshot("initialize scene")
	}

	// Play music
	rl.PlayMusicStream(music)
}

//
// Update
//

func updateGame(dt float64) {
	gameTime += dt
	deltaTime = dt

	// Toggle zoom
	if rl.IsKeyPressed(rl.KEY_Z) {
		if gameCameraZoomTarget == 0.7 {
			gameCameraZoomTarget = 1.2
		} else {
			gameCameraZoomTarget = 0.7
		}
	}
	{
		rate := 14.0
		smoothing := 1 - Pow(2, -rate*deltaTime)
		gameCameraZoom = gameCameraZoom + smoothing*(gameCameraZoomTarget-gameCameraZoom)
		gameCameraSize = gameCameraBaseSize.Scale(gameCameraZoom)
	}

	// Update up direction and clear ground normals
	Each(func(ent Entity, grav *Gravity, lay *Layout) {
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
			if player.JumpsRemaining > 0 && gameTime-player.lastJumpTime > playerJumpCooldown {
				tangentVel := vel.Vel.Subtract(up.Up.Scale(vel.Vel.DotProduct(up.Up)))
				vel.Vel = tangentVel.Add(up.Up.Scale(playerJumpStrength))
				player.lastJumpTime = gameTime
				player.JumpsRemaining--
			}
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
		if collisionShape := GetComponent[CollisionShape](ent); collisionShape != nil {
			Each(func(planetEnt Entity, planet *Planet, planetLay *Layout) {
				nVerts := len(planet.Verts)
				for i, localVertPos := range planet.Verts {
					planetSegmentCapsule := Capsule{
						A:      planetLay.Pos.Add(localVertPos),
						B:      planetLay.Pos.Add(planet.Verts[(i+1)%nVerts]),
						Radius: planetSegmentThickness,
					}

					// Calculate intersection
					poly := Polygon{}
					poly.Count = len(collisionShape.Verts)
					for i := 0; i < poly.Count; i++ {
						poly.Verts[i] = collisionShape.Verts[i]
					}
					poly.CalculateNormals()
					in := IntersectCapsulePolygon(planetSegmentCapsule, &poly, lay.Pos, lay.Rot)
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
							if player := GetComponent[Player](ent); player != nil && player.JumpsRemaining <= 0 {
								player.JumpsRemaining = 2
							}
						}
					}
				}
			})
		}
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

	// Update beam
	Each(func(playerEnt Entity, player *Player, lay *Layout) {
		if rl.IsMouseButtonDown(rl.MOUSE_BUTTON_LEFT) && !player.BuildUIMouseOver {
			start := lay.Pos
			reticlePos := rl.GetScreenToWorld2D(rl.GetMousePosition(), gameCamera)
			reticleDelta := reticlePos.Subtract(start)
			endDist := reticleDelta.Length()
			reticleDir := reticleDelta.Scale(1 / endDist)
			ray := Ray{
				Position:            start,
				NormalizedDirection: reticleDir,
				Length:              endDist,
			}
			hitResourceEnt := NullEntity

			// Check for intersection with planet terrain
			planetSegmentThickness := 0.04
			Each(func(planetEnt Entity, planet *Planet, planetLay *Layout) {
				nVerts := len(planet.Verts)
				for i, localVertPos := range planet.Verts {
					planetSegmentCapsule := Capsule{
						A:      planetLay.Pos.Add(localVertPos),
						B:      planetLay.Pos.Add(planet.Verts[(i+1)%nVerts]),
						Radius: planetSegmentThickness,
					}
					result := RaycastCapsule(ray, planetSegmentCapsule)
					if result.Hit && result.Distance < endDist {
						ray.Length = result.Distance
					}
				}
			})

			// Check for intersection with resources
			Each(func(resourceEnt Entity, resource *Resource, resourceLay *Layout) {
				resourceType := &resourceTypes[resource.TypeId]

				texture := resourceType.texture
				texSize := Vec2{float64(texture.Width), float64(texture.Height)}
				polySize := texSize.Scale(spriteScale).Multiply(Vec2{0.4, 0.8})

				poly := Polygon{}
				poly.Count = 4
				poly.Verts[0] = Vec2{-0.5 * polySize.X, -polySize.Y}
				poly.Verts[1] = Vec2{0.5 * polySize.X, -polySize.Y}
				poly.Verts[2] = Vec2{0.5 * polySize.X, 0}
				poly.Verts[3] = Vec2{-0.5 * polySize.X, 0}
				poly.CalculateNormals()

				result := RaycastPolygon(ray, &poly, resourceLay.Pos, resourceLay.Rot)
				if result.Hit && result.Distance < endDist {
					ray.Length = result.Distance
					hitResourceEnt = resourceEnt
				}
			})

			// Update beam state
			if !player.BeamOn {
				player.BeamOn = true
				player.BeamTimeTillDamage = beamDamagePeriod
			}
			player.BeamEnd = ray.Position.Add(ray.NormalizedDirection.Scale(ray.Length))
			player.BeamTimeSinceStart += deltaTime
			player.BeamTimeTillDamage -= deltaTime
			if player.BeamTimeTillDamage < 0 {
				// Check if we hit a resource
				if hitResourceEnt != NullEntity {
					if resource := GetComponent[Resource](hitResourceEnt); resource != nil {
						// Damage resource a bit
						prevHealth := resource.Health
						resource.Health = Max(0, resource.Health-beamDamage)
						damageDone := prevHealth - resource.Health
						resourceDamaged := AddComponent(hitResourceEnt, ResourceDamaged{})
						resourceDamaged.lastDamageTime = gameTime

						// Consume elements from it
						resourceType := &resourceTypes[resource.TypeId]
						for _, elementAmount := range resourceType.ElementAmounts {
							damagePerAmount := resourceType.Health / Max(1, elementAmount.Amount/2)
							if damagePerAmount > 0 {
								amountGained := prevHealth/damagePerAmount - resource.Health/damagePerAmount
								player.ElementAmounts[elementAmount.TypeId] += amountGained
								if resource.Health == 0 {
									remainingAmount := elementAmount.Amount - resourceType.Health/damagePerAmount
									player.ElementAmounts[elementAmount.TypeId] += remainingAmount
								}
							} else {
								amountPerDamage := (elementAmount.Amount / resourceType.Health) / 2
								player.ElementAmounts[elementAmount.TypeId] += damageDone * amountPerDamage
								if resource.Health == 0 {
									remainingAmount := elementAmount.Amount - (amountPerDamage * resourceType.Health)
									player.ElementAmounts[elementAmount.TypeId] += remainingAmount
								}
							}
						}

						// Destroy resource if it's out of health
						if resource.Health == 0 {
							DestroyEntity(hitResourceEnt)
						}

						// Play damage sound
						if unitRandom() < 0.7 {
							rl.SetSoundVolume(hitSound1, 0.9*1.8)
							rl.SetSoundPitch(hitSound1, 0.60+0.24*unitRandom())
							rl.PlaySound(hitSound1)
						} else {
							rl.SetSoundVolume(hitSound2, 0.6*1.8)
							rl.SetSoundPitch(hitSound2, 0.60+0.28*unitRandom())
							rl.PlaySound(hitSound2)
						}
					}
				}
				player.BeamTimeTillDamage += beamDamagePeriod
			}

			// Check if we should flip player sprite
			delta := player.BeamEnd.Subtract(lay.Pos)
			playerDir := Vec2{Cos(lay.Rot), Sin(lay.Rot)}
			player.FlipH = playerDir.DotProduct(delta) < 0

			// Play sound
			if !rl.IsMusicStreamPlaying(laserSound) {
				rl.SetMusicVolume(laserSound, 0.9)
				rl.PlayMusicStream(laserSound)
			}
		} else {
			player.BeamOn = false
			player.BeamTimeSinceStart = 0
			rl.StopMusicStream(laserSound)
		}
	})

	// Wobble damaged resources
	Each(func(ent Entity, resourceDamaged *ResourceDamaged, resource *Resource) {
		resourceDamaged.time += deltaTime
		if gameTime-resourceDamaged.lastDamageTime < 0.3 {
			height := spriteScale * float64(resourceTypes[resource.TypeId].texture.Height)
			resourceDamaged.rotDeviation = (Pi / 60) * Sin(10*2*Pi*resourceDamaged.time) / height
		} else {
			rate := 420.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			resourceDamaged.rotDeviation *= smoothing
		}
	})

	// Update build UI
	Each(func(ent Entity, player *Player) {
		// Toggle with right click
		if rl.IsMouseButtonPressed(rl.MOUSE_BUTTON_RIGHT) {
			player.BuildUIEnabled = !player.BuildUIEnabled
			if player.BuildUIEnabled {
				player.BuildUIPos = rl.GetScreenToWorld2D(rl.GetMousePosition(), gameCamera)
			}
		}

		// Reset mouseover state on click release
		if rl.IsMouseButtonReleased(rl.MOUSE_BUTTON_LEFT) || rl.IsMouseButtonReleased(rl.MOUSE_BUTTON_RIGHT) {
			player.BuildUIMouseOver = false
		}

		// Build resource if selected
		if player.BuildUISelectedTypeId >= 0 {
			resourceTypeId := player.BuildUISelectedTypeId
			player.BuildUISelectedTypeId = -1
			resourceType := &resourceTypes[resourceTypeId]

			canBuild := true
			for _, requiredElementAmount := range resourceType.ElementAmounts {
				if requiredElementAmount.Amount > player.ElementAmounts[requiredElementAmount.TypeId] {
					canBuild = false
				}
			}
			if canBuild {
				for _, requiredElementAmount := range resourceType.ElementAmounts {
					player.ElementAmounts[requiredElementAmount.TypeId] -= requiredElementAmount.Amount
				}
				buildingEnt := CreateEntity(
					Layout{
						Pos: player.BuildUIPos,
						Rot: -gameCamera.Rotation * Pi / 180,
					},
					Resource{
						TypeId: resourceTypeId,
						Health: resourceType.Health,
					},
					Velocity{},
					Gravity{},
				)
				if len(resourceType.CollisionShapeVerts) > 0 {
					AddComponent(buildingEnt, CollisionShape{
						Verts: resourceType.CollisionShapeVerts,
					})
				}
				if resourceType.Name == "building_refiner" {
					AddComponent(buildingEnt, Refiner{})
				}
			} else {
				// TODO: Play "can't build" sound
			}
		}
	})

	// Stop moving resources that have fallen to the ground
	Each(func(ent Entity, vel *Velocity, resource *Resource, up *Up) {
		if len(up.GroundNormals) > 0 {
			rl.PlaySound(resourceHitGroundSound)
			RemoveComponent[Up](ent)
			RemoveComponent[Gravity](ent)
			RemoveComponent[Velocity](ent)
		}
	})

	// Update refiners
	Each(func(refinerEnt Entity, refiner *Refiner, resource *Resource, refinerLay *Layout) {
		if refiner.CarbonAmount >= refinerCarbonPerFuel {
			// On
			resource.Frame = 0
			refiner.TimeTillNextRefinement -= dt
			if refiner.TimeTillNextRefinement <= 0 {
				refiner.TimeTillNextRefinement += refinerRefinmentPeriod
				refiner.CarbonAmount -= refinerCarbonPerFuel
				refiner.FuelAmount += 1
			}
		} else {
			// Off
			resource.Frame = 1
			refiner.TimeTillNextRefinement = refinerRefinmentPeriod
		}

		// Interaction
		Each(func(playerEnt Entity, player *Player, playerLay *Layout) {
			minDist := 2.0
			if playerLay.Pos.Subtract(refinerLay.Pos).LengthSqr() < minDist*minDist {
				if refiner.FuelAmount > 0 {
					hint := AddComponent(refinerEnt, InteractionHint{})
					hint.Interactable = true
					hint.Message = rl.TextFormat("take %d fuel", refiner.FuelAmount)

					if rl.IsKeyPressed(rl.KEY_E) { // Take all fuel
						player.ElementAmounts[FuelElement] += refiner.FuelAmount
						refiner.FuelAmount = 0
					}
				} else if refiner.CarbonAmount < 100 {
					if player.ElementAmounts[CarbonElement] > 0 {
						hint := AddComponent(refinerEnt, InteractionHint{})
						hint.Interactable = true
						amountAddable := Min(player.ElementAmounts[CarbonElement], 100-refiner.CarbonAmount)
						hint.Message = rl.TextFormat("add %d/100 carbon", amountAddable)

						if rl.IsKeyPressed(rl.KEY_E) {
							refiner.CarbonAmount += amountAddable
							player.ElementAmounts[CarbonElement] -= amountAddable
						}
					} else {
						hint := AddComponent(refinerEnt, InteractionHint{})
						hint.Interactable = false
						hint.Message = "needs carbon"
					}
				} else {
					hint := AddComponent(refinerEnt, InteractionHint{})
					hint.Interactable = false
					hint.Message = "carbon 100/100 full"
				}
			} else {
				RemoveComponent[InteractionHint](refinerEnt)
			}
		})
	})

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

	// Update musics
	rl.SetMusicVolume(music, 0.5)
	rl.UpdateMusicStream(music)
	rl.UpdateMusicStream(laserSound)
}

//
// Draw
//

//gx:extern getAssetPath
func getAssetPath(assetName string) string

var whiteTexture = rl.LoadTextureFromImage(rl.GenImageColor(1, 1, rl.Color{0xff, 0xff, 0xff, 0xff}))

var bitsTextureBasic = rl.LoadTexture(getAssetPath("planet_surface_bits_basic.png"))

var playerTexture = rl.LoadTexture(getAssetPath("player.png"))

var beamSheetTexture = func() rl.Texture {
	result := rl.LoadTexture(getAssetPath("player_beam.png"))
	rl.SetTextureWrap(result, rl.TEXTURE_WRAP_REPEAT)
	return result
}()
var numBeamSheetFrames = 3
var beamSheetFramesPerSecond = 16.0

var beamTipTexture = rl.LoadTexture(getAssetPath("player_beam_tip.png"))
var beamTipCoreTexture = rl.LoadTexture(getAssetPath("player_beam_tip_core.png"))

var reticleTexture = rl.LoadTexture(getAssetPath("cursor.png"))

var elementFrameTexture = rl.LoadTexture(getAssetPath("element_frame.png"))

var buildUIFrameTexture = rl.LoadTexture(getAssetPath("build_interface.png"))

var interactionHintTexture = rl.LoadTexture(getAssetPath("interaction_hint.png"))

func drawGame() {
	rl.ClearBackground(rl.Color{0x10, 0x14, 0x1f, 0xff})

	// Resources
	Each(func(ent Entity, resource *Resource, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rot := lay.Rot
		if resourceDamaged := GetComponent[ResourceDamaged](ent); resourceDamaged != nil {
			rot += resourceDamaged.rotDeviation
		}
		rl.Rotatef(rot*180/Pi, 0, 0, 1)

		resourceType := &resourceTypes[resource.TypeId]
		texture := resourceType.texture
		texSize := Vec2{
			X: float64(texture.Width),
			Y: float64(texture.Height) / float64(resourceType.NumFrames),
		}
		texSource := rl.Rectangle{
			X:      0,
			Y:      float64(resource.Frame) * texSize.Y,
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

	// Beam
	Each(func(ent Entity, player *Player, lay *Layout) {
		if player.BeamOn {
			rl.PushMatrix()
			rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
			delta := player.BeamEnd.Subtract(lay.Pos)
			angle := Atan2(delta.Y, delta.X)
			rl.Rotatef(angle*180/Pi, 0, 0, 1)

			// Green: rl.Color{0xa8, 0xca, 0x58, 0xff}
			// Dark magenta: rl.Color{0x7a, 0x36, 0x7b, 0xff}
			// Light magenta: rl.Color{0xa2, 0x3e, 0x8c, 0xff}
			// Dark blue: rl.Color{0x4f, 0x8f, 0xba, 0xff}
			// Light blue: rl.Color{0x73, 0xbe, 0xd3, 0xff}
			color := rl.Color{0xa8, 0xca, 0x58, 0xff}

			beamLength := delta.Length()
			if player.BeamTimeSinceStart < beamZapTime {
				beamLength *= player.BeamTimeSinceStart / beamZapTime
			}

			// Body
			{
				texHeight := float64(beamSheetTexture.Height)
				texSource := rl.Rectangle{
					X:      0,
					Y:      0,
					Width:  beamLength / spriteScale,
					Height: texHeight / float64(numBeamSheetFrames),
				}
				texSource.Y = float64(Floor(player.BeamTimeSinceStart*beamSheetFramesPerSecond)) * texSource.Height
				destHeight := spriteScale * texSource.Height
				texDest := rl.Rectangle{
					X:      0,
					Y:      -0.5 * destHeight,
					Width:  beamLength,
					Height: destHeight,
				}
				rl.DrawTexturePro(beamSheetTexture, texSource, texDest, Vec2{0, 0}, 0, color)
			}

			// Tip
			drawBeamTip := func(tex rl.Texture, color rl.Color) {
				texWidth := float64(tex.Width)
				texHeight := float64(tex.Height)
				numFrames := int(texHeight / texWidth)
				texSource := rl.Rectangle{
					X:      0,
					Y:      0,
					Width:  texWidth,
					Height: texWidth,
				}
				texSource.Y = float64(Floor(player.BeamTimeSinceStart*beamSheetFramesPerSecond)%numFrames) * texSource.Height
				destWidth := spriteScale * texWidth
				texDest := rl.Rectangle{
					X:      beamLength - 0.5*destWidth,
					Y:      -0.5 * destWidth,
					Width:  destWidth,
					Height: destWidth,
				}
				rl.DrawTexturePro(tex, texSource, texDest, Vec2{0, 0}, 0, color)
			}
			drawBeamTip(beamTipTexture, color)
			rl.BeginBlendMode(rl.BLEND_ADDITIVE)
			drawBeamTip(beamTipCoreTexture, rl.Color{0xff, 0xff, 0xff, 0x42})
			rl.EndBlendMode()

			rl.PopMatrix()
		}
	})

	// Player
	Each(func(ent Entity, player *Player, lay *Layout) {
		rl.PushMatrix()
		rl.Translatef(lay.Pos.X, lay.Pos.Y, 0)
		rl.Rotatef(lay.Rot*180/Pi, 0, 0, 1)

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

		// Terrain
		rl.CheckRenderBatchLimit(4 * (nVerts - 1))
		rl.SetTexture(whiteTexture.Id)
		rl.Begin(rl.Quads)
		for vertIndex, vertPos := range planet.Verts {
			drawTriangleToSegment := func(a, b Vec2) {
				rl.Color4ub(0x15, 0x1d, 0x28, 0xff)
				rl.Vertex2f(0, 0)
				rl.Vertex2f(0, 0)
				rl.Vertex2f(b.X, b.Y)
				rl.Vertex2f(a.X, a.Y)
			}
			drawTriangleToSegment(vertPos, planet.Verts[(vertIndex+1)%nVerts])
		}
		rl.End()

		// Bits
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

				rl.DrawRectangleV(Vec2{}, Vec2{}, rl.Color{}) // <-- Fixes a bug with `DrawTexturePro` below...
				rl.DrawTexturePro(bitTex, bitTexSource, bitTexDest, bitTexOrigin, bit.Rot, rl.Color{0x4d, 0x2b, 0x32, 0xff})
			}
		}

		rl.PopMatrix()
	})

	// HUD
	Each(func(ent Entity, player *Player) {
		// Inventory icons
		{
			rl.PushMatrix()
			worldCameraTopLeft := rl.GetScreenToWorld2D(Vec2{0, 0}, gameCamera)
			rl.Translatef(worldCameraTopLeft.X, worldCameraTopLeft.Y, 0)
			rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
			rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

			iconSize := float64(elementTypes[0].iconTexture.Width)
			iconScreenMargin := 0.5 * iconSize
			iconPos := Vec2{iconScreenMargin, iconScreenMargin}
			for typeId, amount := range player.ElementAmounts {
				elementType := &elementTypes[typeId]
				tex := elementType.iconTexture

				rl.DrawTextureEx(tex, iconPos, 0, 1, rl.White)
				rl.DrawTextureEx(elementFrameTexture, iconPos.SubtractValue(1), 0, 1, rl.White)

				textPos := iconPos.Add(Vec2{0, 1.25 * iconSize})
				fontSize := 0.4 * iconSize
				rl.DrawTextPro(
					rl.GetFontDefault(),
					rl.TextFormat("%d", amount),
					textPos,
					Vec2{0, 0},
					0,
					fontSize,
					1.0,
					rl.Color{0x81, 0x97, 0x96, 0xff},
				)

				iconPos.X += 1.375 * iconSize
			}

			rl.PopMatrix()
		}

		// Build UI
		if player.BuildUIEnabled {
			rl.PushMatrix()
			rl.Translatef(player.BuildUIPos.X, player.BuildUIPos.Y, 0)
			rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
			scale := 2 * gameCameraZoom * spriteScale
			rl.Scalef(scale, scale, 1)

			frameWidth := float64(buildUIFrameTexture.Width)
			frameHeight := float64(buildUIFrameTexture.Height)
			frameOffset := Vec2{0.5 * frameWidth, 1.1 * frameHeight}
			rl.Translatef(-frameOffset.X, -frameOffset.Y, 0)
			rl.DrawTextureEx(buildUIFrameTexture, Vec2{0, 0}, 0, 1, rl.White)

			mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), gameCamera)
			mouseFramePos := mouseWorldPos.
				Subtract(player.BuildUIPos).
				Rotate(gameCamera.Rotation * Pi / 180).
				Scale(1 / scale).
				Add(frameOffset)
			rl.DrawCircleV(mouseFramePos, 1, rl.Red)
			player.BuildUIMouseOver = mouseFramePos.X >= 0 && mouseFramePos.Y >= 0 &&
				mouseFramePos.X <= frameWidth && mouseFramePos.Y <= (frameHeight-8)

			iconFrameMargin := 3.0
			iconSpacing := 3.0
			iconPos := Vec2{iconFrameMargin, iconFrameMargin}
			for typeId, resourceType := range resourceTypes {
				if resourceType.Buildable {
					tex := resourceType.iconTexture
					iconSize := float64(tex.Width)
					rl.DrawTextureEx(tex, iconPos, 0, 1, rl.White)

					if mouseFramePos.X >= iconPos.X && mouseFramePos.Y >= iconPos.Y &&
						mouseFramePos.X <= iconSize && mouseFramePos.Y <= iconSize &&
						rl.IsMouseButtonPressed(rl.MOUSE_BUTTON_LEFT) {
						player.BuildUISelectedTypeId = ResourceTypeId(typeId)
					}

					iconPos.X += iconSize + iconSpacing
				}
			}

			if rl.IsMouseButtonPressed(rl.MOUSE_BUTTON_LEFT) {
				player.BuildUIEnabled = false
			}

			rl.PopMatrix()
		}
	})

	Each(func(ent Entity, interactionHint *InteractionHint, lay *Layout) {
		rl.PushMatrix()
		hintPos := lay.Pos.Add(Vec2{0, interactionHint.VerticalOffset}.Rotate(lay.Rot))
		rl.Translatef(hintPos.X, hintPos.Y, 0)
		rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
		scale := 2 * gameCameraZoom * spriteScale
		rl.Scalef(scale, scale, 1)

		texHeight := float64(interactionHintTexture.Height)

		textHeight := int(0.8 * texHeight)
		frameWidth := float64(rl.MeasureTextEx(rl.GetFontDefault(), interactionHint.Message, float64(textHeight), 1.0).X)
		if interactionHint.Interactable {
			frameWidth += 1.4 * texHeight
		}
		rl.Translatef(-0.5*frameWidth, 0, 0)

		if interactionHint.Interactable {
			rl.DrawTextureEx(interactionHintTexture, Vec2{0, -0.5 * texHeight}, 0, 1, rl.White)
			rl.Translatef(1.4*texHeight, 0, 0)
		}
		rl.DrawTextPro(
			rl.GetFontDefault(),
			interactionHint.Message,
			Vec2{0, -0.5 * float64(textHeight)},
			Vec2{0, 0},
			0,
			float64(textHeight),
			1.0,
			rl.Color{0x81, 0x97, 0x96, 0xff},
		)

		rl.PopMatrix()
	})

	// Reticle
	{
		reticleScale := gameCameraZoom * spriteScale
		reticlePos := rl.GetScreenToWorld2D(rl.GetMousePosition(), gameCamera)
		reticleWidth := float64(reticleTexture.Width) * reticleScale
		reticleHeight := float64(reticleTexture.Height) * reticleScale
		reticleTopLeft := reticlePos.Subtract(Vec2{reticleWidth, reticleHeight}.Scale(0.5))
		rl.DrawTextureEx(reticleTexture, reticleTopLeft, 0, reticleScale, rl.Color{0x81, 0x97, 0x96, 0xff})
	}
}
