package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
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

var editAllowed = false

var surfaceFrictionDecel = 25.0
var atmosphereFrictionDecel = 18.0

var playerSize = Vec2{1.5, 1}

var playerJumpStrength = 14.0
var playerHorizontalControlsAccel = 17.0
var playerMinimumHorizontalSpeedForFriction = 12.0
var playerJumpCooldown = 0.1

var playerLiftoffFuelRequired = 40
var playerFlyingLiftoffInitialAccel = 3.2
var playerFlyingLiftoffJerk = 5.0
var playerFlyingMaxAccel = 4.0
var playerFlyingMaxSpeed = 14.0
var playerFlyingAngAccel = 6.0
var playerFlyingAngDecel = 16.0
var playerFlyingMaxAngSpeed = 1.6
var playerFlyingGravityStrengthMultiplier = 0.3

var planetSegmentThickness = 0.4

var spriteScale = playerSize.X / float64(playerTexture.Width)

var beamZapTime = 0.06
var beamDamagePeriod = 0.2
var beamDamage = 1

var refinerCarbonPerFuel = 5
var refinerCarbonCapacity = 100
var refinerRefinmentPeriod = 5.0
var refinerFuelPerRefinement = 2

var transmissionTowerAntimatterRequired = 5
var transmissionTowerSiliconRequired = 20

var musicFadeSpeed = 0.3

//
// Sounds
//

type MusicTrack struct {
	music     rl.Music
	volume    float64
	maxVolume float64 `default:"0.5"`
}

var musicTracks = []MusicTrack{
	{
		music: rl.LoadMusicStream(getAssetPath("music_menu.ogg")),
	},
	{
		music: rl.LoadMusicStream(getAssetPath("music_1.ogg")),
	},
	{
		music:     rl.LoadMusicStream(getAssetPath("music_2.ogg")),
		maxVolume: 0.7,
	},
}

var musicActiveTrackIndex = 0

var laserSound = rl.LoadMusicStream(getAssetPath("sfx_laser_on.ogg")) // Music so it loops

var hitSound1 = rl.LoadSound(getAssetPath("sfx_hit_1.wav"))
var hitSound2 = rl.LoadSound(getAssetPath("sfx_hit_2.wav"))

var resourceHitGroundSound = rl.LoadSound(getAssetPath("sfx_vehicle_collision.ogg"))

var buildingSucceededSound = rl.LoadSound(getAssetPath("sfx_building_succeeded.ogg"))
var buildingFailedSound = rl.LoadSound(getAssetPath("sfx_building_failed.ogg"))

var winSound = rl.LoadSound(getAssetPath("win.ogg"))

//var liftoffChargeSound = rl.LoadSound(getAssetPath("sfx_liftoff_charge.ogg"))
var liftoffSound = rl.LoadSound(getAssetPath("sfx_liftoff.ogg"))

//var flyingAccelerationSound = rl.LoadMusicStream(getAssetPath("sfx_flying_acceleration.ogg"))

var supernovaSound = rl.LoadSound(getAssetPath("sfx_supernova.ogg"))

//
// Init
//

func unitRandom() float64 {
	return float64(rl.GetRandomValue(0, 100)) / 100.0
}

type FrequencyBand struct {
	Frequency, Amplitude float64
}

type GeneratePlanetTerrainParams struct {
	FrequencyBands []FrequencyBand
}

func generatePlanetTerrain(ent Entity, params GeneratePlanetTerrainParams) {
	planet := GetComponent[Planet](ent)

	// Generation parameters
	segmentLength := 1.5 * playerSize.X
	resolution := 2 * Pi * planet.BaseRadius / segmentLength
	if len(params.FrequencyBands) == 0 {
		params.FrequencyBands = []FrequencyBand{
			{Frequency: 0.003, Amplitude: 0.2 * planet.BaseRadius},
			{Frequency: 0.015, Amplitude: 0.015 * planet.BaseRadius},
			{Frequency: 0.060, Amplitude: 0.015 * planet.BaseRadius},
		}
	}

	// Generate heights and vertices
	angleStep := 2 * Pi / resolution
	for angle := 0.0; angle < 2*Pi; angle += angleStep {
		// Height
		height := planet.BaseRadius
		for _, band := range params.FrequencyBands {
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

	// Smooth out notch among last vertices
	nVerts := len(planet.Verts)
	lastVert := planet.Verts[nVerts-1]
	secondVert := planet.Verts[1]
	planet.Verts[0] = planet.Verts[0].Lerp(lastVert.Lerp(secondVert, 0.5), 0.5)
}

var nextResourceDrawOrder = 0

func createResource(typeId ResourceTypeId, pos Vec2, rot float64, flipH bool) Entity {
	ent := CreateEntity(
		Layout{
			Pos: pos,
			Rot: rot,
		},
		Resource{
			TypeId:    typeId,
			DrawOrder: nextResourceDrawOrder,
			FlipH:     flipH,
			Health:    resourceTypes[typeId].Health,
		},
	)
	nextResourceDrawOrder++
	return ent
}

type GenerateResourcesParams struct {
	TypeName       string
	TypeId         ResourceTypeId
	Planet         Entity
	Resolution     int `default:"16"`
	FrequencyBands []FrequencyBand
	Exponent       float64 `default:"6"`
	Thinning       float64 `default:"1"`
}

func generateResources(params GenerateResourcesParams) {
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

			if unitRandom() < probability {
				upDir := pos.Normalize()
				dir := upDir.Lerp(edgeDir, 0.8)
				rot := Atan2(dir.X, -dir.Y) + Pi/9*(unitRandom()-0.5)
				rotDir := Vec2{Cos(rot), Sin(rot)}

				verticalOffsetDelta := resourceType.VerticalOffsetVariance * unitRandom()
				verticalOffsetDir := Vec2{rotDir.Y, -rotDir.X}
				pos = pos.Add(verticalOffsetDir.Scale(resourceType.VerticalOffset + verticalOffsetDelta))

				createResource(params.TypeId, pos, rot, unitRandom() < 0.5)
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

func initGameplayScene() {
	musicActiveTrackIndex = 1

	rl.SetRandomSeed(1024)

	// Fungal planet
	homePlanetPos := Vec2{0, 24}
	homePlanetRadius := 64.0
	homePlanetEnt := CreateEntity(
		Layout{
			Pos: homePlanetPos,
		},
		Planet{
			BaseRadius:       homePlanetRadius,
			AtmosphereRadius: 1.8 * homePlanetRadius,
			InnerColor:       rl.Color{0x15, 0x1d, 0x28, 0xff},
			BitsColor:        rl.Color{0x4d, 0x2b, 0x32, 0xff},
			AtmosphereColor:  rl.Color{0x10, 0x14, 0x1f, 0xff},
		},
	)
	generatePlanetTerrain(homePlanetEnt, GeneratePlanetTerrainParams{})
	homePlanet := GetComponent[Planet](homePlanetEnt)
	generateResources(GenerateResourcesParams{
		TypeName: "rock_large",
		Planet:   homePlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.5},
			{Frequency: 3, Amplitude: 0.2},
		},
		Exponent: 1,
		Thinning: 0.001,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_medium",
		Planet:   homePlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.5},
			{Frequency: 3, Amplitude: 0.4},
		},
		Exponent: 1,
		Thinning: 0.015,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "fungus_giant",
		Planet:   homePlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Exponent: 1,
		Thinning: 0.02,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "fungus_tiny",
		Planet:   homePlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Thinning: 0.6,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "sprout_tiny",
		Planet:   homePlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Exponent: 2,
	})

	// Fungal planet sibling 1 (medium size)
	homeSibling1PlanetPos := Vec2{100, 24}
	homeSibling1PlanetRadius := 0.3 * homePlanetRadius
	homeSibling1PlanetEnt := CreateEntity(
		Layout{
			Pos: homeSibling1PlanetPos,
		},
		Planet{
			BaseRadius:       homeSibling1PlanetRadius,
			AtmosphereRadius: 1.5 * homeSibling1PlanetRadius,
			InnerColor:       homePlanet.InnerColor,
			BitsColor:        homePlanet.BitsColor,
			AtmosphereColor:  homePlanet.AtmosphereColor,
		},
	)
	generatePlanetTerrain(homeSibling1PlanetEnt, GeneratePlanetTerrainParams{})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_medium",
		Planet:   homeSibling1PlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.5},
			{Frequency: 3, Amplitude: 0.4},
		},
		Exponent: 1,
		Thinning: 0.015,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "fungus_tiny",
		Planet:   homeSibling1PlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Thinning: 0.6,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "sprout_tiny",
		Planet:   homeSibling1PlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Exponent: 2,
	})

	// Fungal planet sibling 2 (very small)
	homeSibling2PlanetPos := Vec2{80, 53}
	homeSibling2PlanetRadius := 0.1 * homePlanetRadius
	homeSibling2PlanetEnt := CreateEntity(
		Layout{
			Pos: homeSibling2PlanetPos,
		},
		Planet{
			BaseRadius:       homeSibling2PlanetRadius,
			AtmosphereRadius: 1.5 * homeSibling2PlanetRadius,
			InnerColor:       homePlanet.InnerColor,
			BitsColor:        homePlanet.BitsColor,
			AtmosphereColor:  homePlanet.AtmosphereColor,
		},
	)
	generatePlanetTerrain(homeSibling2PlanetEnt, GeneratePlanetTerrainParams{})
	generateResources(GenerateResourcesParams{
		TypeName: "fungus_tiny",
		Planet:   homeSibling2PlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Thinning: 0.6,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "sprout_tiny",
		Planet:   homeSibling2PlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Exponent: 2,
		Thinning: 0.8,
	})

	// Fungal planet antiplants
	antiplantResourceTypeId := resourceTypeIdForName("antiplant")
	createResource(
		antiplantResourceTypeId,
		Vec2{-42.84, -21.56},
		-0.7,
		unitRandom() < 0.5,
	)
	createResource(
		antiplantResourceTypeId,
		Vec2{118.2, 18.42},
		1.35,
		unitRandom() < 0.5,
	)

	// Ending planet
	endingPlanetPos := Vec2{0, -200}
	endingPlanetRadius := 0.8 * homePlanetRadius
	endingPlanetEnt := CreateEntity(
		Layout{
			Pos: endingPlanetPos,
		},
		Planet{
			BaseRadius:       endingPlanetRadius,
			AtmosphereRadius: 1.5 * endingPlanetRadius,
			InnerColor:       rl.Color{0x24, 0x15, 0x27, 0xff},
			BitsColor:        rl.Color{0x7a, 0x36, 0x7b, 0xff},
			AtmosphereColor:  rl.Color{0x10, 0x14, 0x1f, 0xff},
		},
		ArrowTarget{},
		EndingPlanet{},
	)
	generatePlanetTerrain(endingPlanetEnt, GeneratePlanetTerrainParams{
		FrequencyBands: []FrequencyBand{
			{Frequency: 0.003, Amplitude: 0.28 * endingPlanetRadius},
			{Frequency: 0.012, Amplitude: 0.095 * endingPlanetRadius},
			{Frequency: 0.055, Amplitude: 0.020 * endingPlanetRadius},
			{Frequency: 0.102, Amplitude: 0.012 * endingPlanetRadius},
		},
	})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_large",
		Planet:   endingPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 30, Amplitude: 0.5},
			{Frequency: 5, Amplitude: 0.2},
		},
		Exponent: 2,
		Thinning: 0.001,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_medium",
		Planet:   endingPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.5},
			{Frequency: 3, Amplitude: 0.4},
		},
		Exponent: 3,
		Thinning: 0.02,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "antiplant",
		Planet:   endingPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.5},
			{Frequency: 3, Amplitude: 0.4},
		},
		Exponent: 1,
		Thinning: 0.015,
	})
	transmissionTowerResourceTypeId := resourceTypeIdForName("transmission_tower")
	transmissionTowerEnt := createResource(
		transmissionTowerResourceTypeId,
		Vec2{-9.77, -251.381},
		-0.21,
		false,
	)
	AddComponent(transmissionTowerEnt, TransmissionTower{})

	// Player
	//playerPos := Vec2{-9.77, -251.381 - 8} // At transmission tower!
	playerPos := Vec2{-26.66, 82.36}
	playerRot := -2.723
	playerPolySize := playerSize.Subtract(Vec2{2 * planetSegmentThickness, 2.14 * planetSegmentThickness})
	playerEnt := CreateEntity(
		Layout{
			Pos: playerPos,
			Rot: playerRot,
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
			CameraPos:       playerPos,
			CameraRot:       playerRot,
			TimeToSupernova: 7.5 * 60,
			//TimeToSupernova: 3,
		},
	)
	edit.Camera().Target = playerPos
	player := GetComponent[Player](playerEnt)
	player.ElementAmounts[CarbonElement] = 0
	//player.ElementAmounts[CarbonElement] = 3000
	//player.ElementAmounts[SiliconElement] = 3000
	//player.ElementAmounts[FuelElement] = 3000
	//player.ElementAmounts[AntimatterElement] = 3000

	if editAllowed {
		edit.SaveSnapshot("initialize scene")
	}
}

var menuActive = false

func initMenuScene() {
	musicActiveTrackIndex = 0

	menuActive = true

	// Menu planet
	menuPlanetPos := Vec2{0, 0}
	menuPlanetRadius := 30.0
	menuPlanetEnt := CreateEntity(
		Layout{
			Pos: menuPlanetPos,
		},
		Planet{
			BaseRadius:       menuPlanetRadius,
			AtmosphereRadius: 3 * menuPlanetRadius,
			InnerColor:       rl.Color{0x15, 0x1d, 0x28, 0xff},
			BitsColor:        rl.Color{0x4d, 0x2b, 0x32, 0xff},
			AtmosphereColor:  rl.Color{0x10, 0x14, 0x1f, 0xff},
		},
	)
	generatePlanetTerrain(menuPlanetEnt, GeneratePlanetTerrainParams{})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_large",
		Planet:   menuPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 60, Amplitude: 0.2},
			{Frequency: 3, Amplitude: 0.2},
		},
		Exponent: 1,
		Thinning: 0.001,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "rock_medium",
		Planet:   menuPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 42.6, Amplitude: 0.5},
			{Frequency: 5, Amplitude: 0.4},
		},
		Exponent: 1,
		Thinning: 0.015,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "fungus_tiny",
		Planet:   menuPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Thinning: 0.6,
	})
	generateResources(GenerateResourcesParams{
		TypeName: "sprout_tiny",
		Planet:   menuPlanetEnt,
		FrequencyBands: []FrequencyBand{
			{Frequency: 80, Amplitude: 0.5},
			{Frequency: 16, Amplitude: 0.5},
		},
		Exponent: 2,
	})

	// Camera
	gameCamera.Target = Vec2{0, 36}
	gameCamera.Rotation = 180
}

func initGame() {
	rl.HideCursor()

	// Initialize element and resource assets
	for _, elementType := range elementTypes {
		elementType.iconTexture = rl.LoadTexture(getAssetPath(elementType.IconImageName))
	}
	for _, resourceType := range resourceTypes {
		resourceType.texture = rl.LoadTexture(getAssetPath(resourceType.ImageName))
		if resourceType.IconImageName != "" {
			resourceType.iconTexture = rl.LoadTexture(getAssetPath(resourceType.IconImageName))
		}
		if resourceType.DestructionSoundName != "" {
			resourceType.destructionSound = rl.LoadSound(getAssetPath(resourceType.DestructionSoundName))
		}
	}

	// Edit session
	if editAllowed && edit.LoadSession() {
		return
	}

	// Enter menu!
	initMenuScene()
	//initGameplayScene()
}

//
// Update
//

var restartGameplay = false

func updateGame(dt float64) {
	gameTime += dt
	deltaTime = dt

	// Menu camera rotation
	if menuActive {
		rotSpeed := 0.03
		rotDelta := rotSpeed * deltaTime
		gameCamera.Target = gameCamera.Target.Rotate(rotDelta)
		gameCamera.Rotation -= rotDelta * 180 / Pi
	}

	// Update camera zoom
	Each(func(ent Entity, player *Player) {
		if player.Transmitting {
			// Slow zoom out
			gameCameraZoom += 0.2 * deltaTime
			gameCameraSize = gameCameraBaseSize.Scale(gameCameraZoom)
		} else {
			// Toggle zoom
			if rl.IsKeyPressed(rl.KEY_Z) {
				if gameCameraZoomTarget == 0.7 {
					gameCameraZoomTarget = 1.2
				} else {
					gameCameraZoomTarget = 0.7
				}
			}

			// Smooth to target zoom
			rate := 14.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			gameCameraZoom = gameCameraZoom + smoothing*(gameCameraZoomTarget-gameCameraZoom)
			gameCameraSize = gameCameraBaseSize.Scale(gameCameraZoom)
		}
	})

	// Clear collision normals
	ClearComponent[CollisionNormals]()

	// Update up direction and grounded state
	Each(func(ent Entity, up *Up, lay *Layout) {
		minSqDist := -1.0
		minDelta := Vec2{0, 0}
		Each(func(ent Entity, planet *Planet, planetLay *Layout) {
			delta := planetLay.Pos.Subtract(lay.Pos)
			sqDist := delta.LengthSqr()
			if minSqDist < 0 || sqDist < minSqDist {
				if sqDist < planet.AtmosphereRadius*planet.AtmosphereRadius {
					minSqDist = sqDist
					minDelta = delta
				}
			}
		})
		if minSqDist > 0 {
			dir := minDelta.Scale(1 / Sqrt(minSqDist))
			up.Up = dir.Negate().Normalize()
		}
		up.grounded = false
	})

	// Grounded controls
	Each(func(ent Entity, player *Player, up *Up, vel *Velocity) {
		if player.Flying {
			return
		}
		RemoveComponent[AngularVelocity](ent)

		// Horizontal
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

		// Jump
		if rl.IsKeyPressed(rl.KEY_W) || rl.IsKeyPressed(rl.KEY_UP) {
			if player.JumpsRemaining > 0 && gameTime-player.lastJumpTime > playerJumpCooldown {
				tangentVel := vel.Vel.Subtract(up.Up.Scale(vel.Vel.DotProduct(up.Up)))
				vel.Vel = tangentVel.Add(up.Up.Scale(playerJumpStrength))
				player.lastJumpTime = gameTime
				player.JumpsRemaining--
			}
		}
	})

	// Flying controls
	playFlyingAccelerationSound := false
	Each(func(ent Entity, player *Player, lay *Layout) {
		if !player.Flying {
			return
		}
		angVel := AddComponent(ent, AngularVelocity{})

		if rl.IsKeyDown(rl.KEY_W) || rl.IsKeyDown(rl.KEY_UP) {
			playFlyingAccelerationSound = true
			player.FlyingAccel += Min(playerFlyingLiftoffJerk*deltaTime, playerFlyingMaxAccel)
		}
		if vel := GetComponent[Velocity](ent); vel != nil {
			// Acceleration / slowdown
			forwardDir := Vec2{0, -1}.Rotate(lay.Rot)
			if rl.IsKeyDown(rl.KEY_W) || rl.IsKeyDown(rl.KEY_UP) {
				playFlyingAccelerationSound = true
				vel.Vel = vel.Vel.Add(forwardDir.Scale(player.FlyingAccel * deltaTime))
			}
			if rl.IsKeyDown(rl.KEY_S) || rl.IsKeyDown(rl.KEY_DOWN) {
				playFlyingAccelerationSound = true
				vel.Vel = vel.Vel.Add(forwardDir.Scale(-player.FlyingAccel * deltaTime))
			}
			if speed := vel.Vel.Length(); speed > playerFlyingMaxSpeed {
				vel.Vel = vel.Vel.Scale(playerFlyingMaxSpeed / speed)
			}
			rate := 400.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			vel.Vel = vel.Vel.Scale(smoothing)

			// Turning
			appliedTurning := false
			if rl.IsKeyDown(rl.KEY_A) || rl.IsKeyDown(rl.KEY_LEFT) {
				angVel.AngVel -= playerFlyingAngAccel * deltaTime
				appliedTurning = true
			}
			if rl.IsKeyDown(rl.KEY_D) || rl.IsKeyDown(rl.KEY_RIGHT) {
				angVel.AngVel += playerFlyingAngAccel * deltaTime
				appliedTurning = true
			}

			// Angular velocity limits
			angSpeed := Abs(angVel.AngVel)
			if !appliedTurning {
				angSpeed = Max(0, angSpeed-playerFlyingAngDecel*deltaTime)
			}
			angSpeed = Min(angSpeed, playerFlyingMaxAngSpeed)
			angVel.AngVel = Sign(angVel.AngVel) * angSpeed
		} else {
			// Still in liftoff
			grav := GetComponent[Gravity](ent)
			if grav == nil || player.FlyingAccel >= 0.85*grav.Strength {
				musicActiveTrackIndex = -1
				AddComponent(ent, Velocity{})
				rl.PlaySound(liftoffSound)
			}
		}
	})
	if playFlyingAccelerationSound {
		//rl.PlayMusicStream(flyingAccelerationSound)
	} else {
		//rl.StopMusicStream(flyingAccelerationSound)
	}

	// Gravity toward planets
	Each(func(ent Entity, grav *Gravity, vel *Velocity, lay *Layout) {
		first := true
		minSurfaceDist := -1.0
		minDist := -1.0
		minDelta := Vec2{0, 0}
		Each(func(ent Entity, planet *Planet, planetLay *Layout) {
			// TODO: Falloff around atmosphere radius
			delta := planetLay.Pos.Subtract(lay.Pos)
			sqDist := delta.LengthSqr()
			if sqDist < planet.AtmosphereRadius*planet.AtmosphereRadius {
				dist := Sqrt(sqDist)
				surfaceDist := dist - planet.BaseRadius
				if first || surfaceDist < minSurfaceDist {
					first = false
					minDist = dist
					minDelta = delta
				}
			}
		})
		if minDist > 0 {
			dir := minDelta.Scale(1 / minDist)
			vel.Vel = vel.Vel.Add(dir.Scale(grav.Strength * deltaTime))
		}
	})

	// Apply velocity
	Each(func(ent Entity, vel *Velocity, lay *Layout) {
		up := GetComponent[Up](ent)

		// Apply velocity
		lay.Pos = lay.Pos.Add(vel.Vel.Scale(deltaTime))

		// Handle collisons
		// TODO: Flying -- shape?
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
						collisionNormals := AddComponent(ent, CollisionNormals{})
						ground := up != nil && in.Normal.DotProduct(up.Up) > 0.2
						collisionNormals.Normals = append(collisionNormals.Normals, CollisionNormal{
							Normal: in.Normal,
							Ground: ground,
						})
						if ground {
							up.grounded = true
							up.lastGroundedTime = gameTime
						}
					}
				}
			})
		}
	})

	// Apply angular velocity
	Each(func(ent Entity, lay *Layout, angVel *AngularVelocity) {
		lay.Rot += angVel.AngVel * deltaTime
		for lay.Rot > 2*Pi {
			lay.Rot -= 2 * Pi
		}
		for lay.Rot < 0 {
			lay.Rot += 2 * Pi
		}
	})

	// Donate jumps if grounded
	Each(func(ent Entity, player *Player, up *Up) {
		if up.grounded {
			player.JumpsRemaining = 2
		}
	})

	// Eject flying if hit ground
	Each(func(ent Entity, player *Player, collisionNormals *CollisionNormals) {
		if player.Flying && len(collisionNormals.Normals) != 0 {
			player.Flying = false
			AddComponent(ent, Up{})
			RemoveComponent[Gravity](ent) // Bring back regular gravity
			AddComponent(ent, Gravity{})
		}
	})

	// Update auto-upright
	Each(func(ent Entity, up *Up, lay *Layout) {
		target := Vec2{0, 0}
		rate := 28.0
		if !up.grounded {
			if gameTime-up.lastGroundedTime < 0.3 {
				rate = 7.0
			}
			target = up.Up
		} else {
			collisionNormals := GetComponent[CollisionNormals](ent)
			count := 0
			for _, normal := range collisionNormals.Normals {
				if normal.Ground {
					target = target.Add(normal.Normal)
					count++
				}
			}
			target = target.Scale(1 / float64(count))
		}
		smoothing := 1 - Pow(2, -rate*deltaTime)
		up.AutoUprightDir = up.AutoUprightDir.Add(target.Subtract(up.AutoUprightDir).Scale(smoothing)).Normalize()
		lay.Rot = Atan2(up.AutoUprightDir.X, -up.AutoUprightDir.Y)
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
	Each(func(ent Entity, fric *ApplySurfaceFriction, collisionNormals *CollisionNormals, vel *Velocity) {
		if !HasComponent[Up](ent) && !HasComponent[DisableFriction](ent) {
			for _, normal := range collisionNormals.Normals {
				normalVel := normal.Normal.Scale(vel.Vel.DotProduct(normal.Normal))
				tangentVel := vel.Vel.Subtract(normalVel)
				if tangentSpeed := tangentVel.Length(); tangentSpeed > 0 {
					tangentDir := tangentVel.Scale(1 / tangentSpeed)
					tangentSpeed = Max(0, tangentSpeed-surfaceFrictionDecel*deltaTime)
					vel.Vel = normalVel.Add(tangentDir.Scale(tangentSpeed))
				}
			}
		}
	})
	ClearComponent[DisableFriction]()
	ClearComponent[ApplySurfaceFriction]()

	// Update beam
	Each(func(playerEnt Entity, player *Player, lay *Layout) {
		if player.TimeToSupernova <= 0 {
			return
		}
		if !player.Flying && rl.IsMouseButtonDown(rl.MOUSE_BUTTON_LEFT) && !player.BuildUIMouseOver {
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
						if resourceType := &resourceTypes[resource.TypeId]; resourceType.Damageable {
							// Damage resource a bit
							prevHealth := resource.Health
							resource.Health = Max(0, resource.Health-beamDamage)
							damageDone := prevHealth - resource.Health
							resourceDamaged := AddComponent(hitResourceEnt, ResourceDamaged{})
							resourceDamaged.lastDamageTime = gameTime

							// Consume elements from it
							for _, elementAmount := range resourceType.ElementAmounts {
								if elementAmount.Amount == 1 {
									if resource.Health == 0 {
										player.ElementAmounts[elementAmount.TypeId] += elementAmount.Amount
									}
								} else {
									damagePerAmount := resourceType.Health / (elementAmount.Amount / 2)
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
							}

							// Destroy resource if it's out of health
							if resource.Health == 0 {
								DestroyEntity(hitResourceEnt)
								rl.PlaySound(resourceType.destructionSound)
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
		// Can't build when flying
		if player.Flying {
			player.BuildUIEnabled = false
			return
		}

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
				buildingEnt := createResource(
					resourceTypeId,
					player.BuildUIPos,
					-gameCamera.Rotation*Pi/180,
					false,
				)
				AddComponent(buildingEnt, Velocity{})
				AddComponent(buildingEnt, Up{})
				AddComponent(buildingEnt, Gravity{})
				if len(resourceType.CollisionShapeVerts) > 0 {
					AddComponent(buildingEnt, CollisionShape{
						Verts: resourceType.CollisionShapeVerts,
					})
				}
				if resourceType.Name == "refiner" {
					AddComponent(buildingEnt, Refiner{})
				} else if resourceType.Name == "launchpad" {
					AddComponent(buildingEnt, Launchpad{})
				}
			} else {
				rl.SetSoundVolume(buildingFailedSound, 0.5)
				rl.PlaySound(buildingFailedSound)
			}
		}
	})

	// Stop moving resources that have fallen to the ground
	Each(func(ent Entity, vel *Velocity, resource *Resource, up *Up) {
		if up.grounded {
			rl.SetSoundVolume(resourceHitGroundSound, 0.8)
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
			refiner.TimeTillNextRefinement -= deltaTime
			if refiner.TimeTillNextRefinement <= 0 {
				refiner.TimeTillNextRefinement += refinerRefinmentPeriod
				refiner.CarbonAmount -= refinerCarbonPerFuel * refinerFuelPerRefinement
				refiner.FuelAmount += refinerFuelPerRefinement
			}
		} else {
			// Off
			resource.Frame = 1
			refiner.TimeTillNextRefinement = refinerRefinmentPeriod
		}

		// Interaction
		Each(func(playerEnt Entity, player *Player, playerLay *Layout) {
			minDist := 2.0
			if !player.Flying && playerLay.Pos.Subtract(refinerLay.Pos).LengthSqr() < minDist*minDist {
				if refiner.FuelAmount > 0 {
					hint := AddComponent(refinerEnt, InteractionHint{})
					hint.Interactable = true
					hint.Message = rl.TextFormat("take %d fuel", refiner.FuelAmount)

					if rl.IsKeyPressed(rl.KEY_E) { // Take all fuel
						player.ElementAmounts[FuelElement] += refiner.FuelAmount
						refiner.FuelAmount = 0
					}
				} else if refiner.CarbonAmount < refinerCarbonCapacity {
					if player.ElementAmounts[CarbonElement] > 0 {
						hint := AddComponent(refinerEnt, InteractionHint{})
						hint.Interactable = true
						amountAddable := Min(player.ElementAmounts[CarbonElement], refinerCarbonCapacity-refiner.CarbonAmount)
						hint.Message = rl.TextFormat("add %d/%d carbon", amountAddable, refinerCarbonCapacity)

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
					hint.Message = rl.TextFormat("carbon %d/%d full", refinerCarbonCapacity, refinerCarbonCapacity)
				}
			} else {
				RemoveComponent[InteractionHint](refinerEnt)
			}
		})
	})

	// Update launchpads
	Each(func(launchpadEnt Entity, launchpad *Launchpad, launchpadLay *Layout) {
		// Interaction
		Each(func(playerEnt Entity, player *Player, playerLay *Layout) {
			minDist := 2.0
			if !player.Flying && playerLay.Pos.Subtract(launchpadLay.Pos).LengthSqr() < minDist*minDist {
				if player.ElementAmounts[FuelElement] >= playerLiftoffFuelRequired {
					hint := AddComponent(launchpadEnt, InteractionHint{})
					hint.Interactable = true
					hint.Message = "launch"

					if rl.IsKeyPressed(rl.KEY_E) {
						// Consume fuel
						player.ElementAmounts[FuelElement] -= playerLiftoffFuelRequired

						// Enter flying mode
						RemoveComponent[Up](playerEnt)
						player.Liftoff = true
						player.Flying = true
						player.FlyingAccel = playerFlyingLiftoffInitialAccel

						// Snap to launchpad
						playerLay.Pos = launchpadLay.Pos.Add(Vec2{0.21, -3}.Rotate(launchpadLay.Rot))
						playerLay.Rot = launchpadLay.Rot
						RemoveComponent[Velocity](playerEnt)

						// Reduce gravity strength
						if grav := GetComponent[Gravity](playerEnt); grav != nil {
							grav.Strength *= playerFlyingGravityStrengthMultiplier
						}
					}
				} else {
					hint := AddComponent(launchpadEnt, InteractionHint{})
					hint.Interactable = false
					hint.Message = rl.TextFormat("needs %d fuel to launch", playerLiftoffFuelRequired)
				}
			} else {
				RemoveComponent[InteractionHint](launchpadEnt)
			}
		})
	})

	// Update transmission tower
	Each(func(transmissionTowerEnt Entity, transmissionTower *TransmissionTower, transmissionTowerLay *Layout) {
		// Interaction
		Each(func(playerEnt Entity, player *Player, playerLay *Layout) {
			minDist := 1.0
			if !player.Transmitting && !player.Flying &&
				playerLay.Pos.Subtract(transmissionTowerLay.Pos).LengthSqr() < minDist*minDist {
				if player.ElementAmounts[AntimatterElement] >= transmissionTowerAntimatterRequired &&
					player.ElementAmounts[SiliconElement] >= transmissionTowerSiliconRequired {
					hint := AddComponent(transmissionTowerEnt, InteractionHint{})
					hint.Interactable = true
					hint.Message = "transmit"

					if rl.IsKeyPressed(rl.KEY_E) {
						// Consume elements
						player.ElementAmounts[AntimatterElement] -= transmissionTowerAntimatterRequired
						player.ElementAmounts[SiliconElement] -= transmissionTowerSiliconRequired

						// Switch to transmitting
						player.Transmitting = true

						// Snap to transmission tower
						playerLay.Pos = transmissionTowerLay.Pos.Add(Vec2{0, -1}.Rotate(transmissionTowerLay.Rot))
						playerLay.Rot = transmissionTowerLay.Rot
						RemoveComponent[Velocity](playerEnt)
						RemoveComponent[Velocity](playerEnt)

						// Win!
						musicActiveTrackIndex = 0
						rl.SetSoundVolume(winSound, 0.4)
						rl.PlaySound(winSound)
					}
				} else {
					hint := AddComponent(transmissionTowerEnt, InteractionHint{})
					hint.Interactable = false
					hint.Message = rl.TextFormat(
						"needs %d silicon, %d antimatter",
						transmissionTowerSiliconRequired,
						transmissionTowerAntimatterRequired,
					)
				}
			} else {
				RemoveComponent[InteractionHint](transmissionTowerEnt)
			}
		})
	})

	// Update transmission effect
	Each(func(ent Entity, player *Player) {
		if player.Transmitting {
			rate := 400.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			player.TransmitWidthScale *= smoothing
		}
	})

	// Update supernova timer
	Each(func(ent Entity, player *Player) {
		player.TimeToSupernova -= deltaTime
		if player.TimeToSupernova <= 0 {
			bloomEnabled = true
			factor := Abs(0.35*player.TimeToSupernova) + 1
			bloomBrightness = factor * factor
			if !rl.IsSoundPlaying(supernovaSound) {
				rl.PlaySound(supernovaSound)
			}
		}
		if player.TimeToSupernova <= -7 {
			restartGameplay = true
		}
	})

	// Ending planet music when entering its atmosphere
	Each(func(ent Entity, player *Player, lay *Layout) {
		if player.Transmitting {
			return
		}
		Each(func(ent Entity, endingPlanet *EndingPlanet, planet *Planet, planetLay *Layout) {
			delta := planetLay.Pos.Subtract(lay.Pos)
			sqDist := delta.LengthSqr()
			if sqDist < planet.AtmosphereRadius*planet.AtmosphereRadius {
				musicActiveTrackIndex = 2
			}
		})
	})

	// Update camera
	Each(func(ent Entity, player *Player, lay *Layout) {
		upOffset := Vec2{0, 0}
		velOffset := Vec2{0, 0}
		if !player.Flying {
			// Offset look-at by smoothed velocity and up direction when not flying
			rate := 160.0
			smoothing := 1 - Pow(2, -rate*deltaTime)
			currVel := Vec2{0, 0}
			if vel := GetComponent[Velocity](ent); vel != nil {
				currVel = vel.Vel
			}
			player.SmoothedVel = player.SmoothedVel.Lerp(currVel, smoothing)
			velOffset = player.SmoothedVel.Scale(0.4 * gameCameraZoom)
			if up := GetComponent[Up](ent); up != nil {
				upOffset = up.Up.Scale(0.9)
				upVelOffset := up.Up.Scale(velOffset.DotProduct(up.Up))
				velOffset = velOffset.Subtract(upVelOffset).Add(upVelOffset.Scale(0.1))
			}
		}
		targetPos := lay.Pos.Add(upOffset).Add(velOffset)

		if !player.CameraInitialized {
			// Immediate snap on initialization
			player.CameraPos = targetPos
			player.CameraRot = lay.Rot
			player.CameraInitialized = true
		} else {
			// Smooth to position
			{
				zoomDelta := 1 - gameCameraZoom
				rateFactor := zoomDelta * zoomDelta * zoomDelta
				rate := 14.0 / (1 - rateFactor)
				smoothing := 1 - Pow(2, -rate*deltaTime)
				player.CameraPos = player.CameraPos.Lerp(targetPos, smoothing)
			}

			// Smooth to rotation when not flying
			if !player.Flying {
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

		// Apply to camera, fix angle
		gameCamera.Target = player.CameraPos
		for player.CameraRot > 2*Pi {
			player.CameraRot -= 2 * Pi
		}
		for player.CameraRot < 0 {
			player.CameraRot += 2 * Pi
		}
		gameCamera.Rotation = -player.CameraRot * 180 / Pi
	})

	// Check for gameplay restart request
	if restartGameplay {
		restartGameplay = false
		bloomEnabled = false
		Each(func(ent Entity) {
			DestroyEntity(ent)
		})
		initGameplayScene()
	}

	// Update music tracks
	for index, track := range musicTracks {
		if index == musicActiveTrackIndex {
			if !rl.IsMusicStreamPlaying(track.music) {
				rl.PlayMusicStream(track.music)
			}
			if track.volume < track.maxVolume {
				track.volume = Min(track.volume+musicFadeSpeed*deltaTime, track.maxVolume)
			}
		} else {
			if track.volume > 0 {
				track.volume = Max(0, track.volume-musicFadeSpeed*deltaTime)
			}
			if track.volume <= 0 && rl.IsMusicStreamPlaying(track.music) {
				rl.StopMusicStream(track.music)
			}
		}
		rl.SetMusicVolume(track.music, track.volume)
		rl.UpdateMusicStream(track.music)
	}

	// Sounds that we want to loop
	rl.UpdateMusicStream(laserSound)
	//rl.UpdateMusicStream(flyingAccelerationSound)
}

//
// Draw
//

//gx:extern getAssetPath
func getAssetPath(assetName string) string

//gx:extern nullptr
var nilString = ""

//gx:extern (const char *)
func cString(s string) string

var whiteTexture = func() rl.Texture {
	image := rl.GenImageColor(1, 1, rl.Color{0xff, 0xff, 0xff, 0xff})
	result := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	return result
}()

var primaryStarTexture = rl.LoadTexture(getAssetPath("primary_star.png"))
var countdownFont = func() rl.Font {
	result := rl.LoadFontEx(getAssetPath("countdown.ttf"), 20, nil, 0)
	rl.SetTextureFilter(result.Texture, rl.TEXTURE_FILTER_POINT)
	return result
}()

var bitsTextureBasic = rl.LoadTexture(getAssetPath("planet_surface_bits_basic.png"))

var playerTexture = rl.LoadTexture(getAssetPath("player.png"))
var playerShipTexture = rl.LoadTexture(getAssetPath("player_ship.png"))

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

var arrowTexture = rl.LoadTexture(getAssetPath("arrow.png"))

var buildUIFrameTexture = rl.LoadTexture(getAssetPath("build_interface.png"))

var interactionHintTexture = rl.LoadTexture(getAssetPath("interaction_hint.png"))

type GlobalHintMessage struct {
	Message string
	Color   rl.Color `default:"{ 0x81, 0x97, 0x96, 0xff }"`
}

var globalHintMessages []GlobalHintMessage

var titleTexture = rl.LoadTexture(getAssetPath("packet_lost_title.png"))
var playButtonTexture = rl.LoadTexture(getAssetPath("play_button.png"))

var screenTexture = rl.LoadRenderTexture(rl.GetScreenWidth(), rl.GetScreenHeight())

var bloomEnabled = false
var bloomBrightness = 1.0
var bloomShader = rl.LoadShader(nilString, getAssetPath("bloom.frag"))

func drawGame() {
	// Begin bloom if enabled
	if bloomEnabled {
		screenWidth := rl.GetScreenWidth()
		screenHeight := rl.GetScreenHeight()
		if screenTexture.Texture.Width != screenWidth || screenTexture.Texture.Height != screenHeight {
			rl.UnloadRenderTexture(screenTexture)
			screenTexture = rl.LoadRenderTexture(screenWidth, screenHeight)
		}
		rl.BeginTextureMode(screenTexture)
	}

	// Begin camera
	if edit.Enabled() {
		rl.BeginMode2D(*edit.Camera())
	} else {
		rl.BeginMode2D(gameCamera)
	}

	// Background color
	rl.ClearBackground(rl.Color{0x09, 0x0a, 0x14, 0xff})

	// Planet atmospheres
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		rl.DrawCircleSector(lay.Pos, planet.AtmosphereRadius, 0, 360, 128, planet.AtmosphereColor)
	})

	// Primary star
	{
		rl.PushMatrix()
		screenCenter := Vec2{float64(rl.GetScreenWidth()), float64(rl.GetScreenHeight())}.Scale(0.5)
		worldCameraCenter := rl.GetScreenToWorld2D(screenCenter, gameCamera).Scale(0.98)
		rl.Translatef(worldCameraCenter.X, worldCameraCenter.Y, 0)
		rl.Scalef(gameCameraZoom, gameCameraZoom, 1)

		pos := Vec2{-5, 3}.Scale(1 / 0.98)
		rl.Translatef(pos.X, pos.Y, 0)

		rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
		texSize := Vec2{float64(primaryStarTexture.Width), float64(primaryStarTexture.Height)}.Scale(2 * spriteScale)
		scale := 1.0
		Each(func(ent Entity, player *Player) {
			if player.TimeToSupernova <= -1 {
				stretch := Abs(player.TimeToSupernova)
				scale = 1 + Pow(Max(0, 0.35*(stretch-1)), 2)

				rl.PushMatrix()
				rl.Scalef(Pow(stretch-1, 4), 0.4, 1)
				rl.DrawTextureEx(primaryStarTexture, texSize.Scale(-0.5), 0, 2*spriteScale, rl.White)
				rl.PopMatrix()
			}
		})
		rl.DrawTextureEx(primaryStarTexture, texSize.Scale(-0.5).Scale(scale), 0, 2*spriteScale*scale, rl.White)
		if scale > 2 {
			rl.DrawCircleV(Vec2{0, 0}, 0.5*texSize.X*scale, rl.Color{0xad, 0x77, 0x57, 0xff})
		}

		//rl.Translatef(0, -0.55*texSize.Y, 0)
		//
		////darkRed := rl.Color{0x4d, 0x2b, 0x32, 0xff}
		//darkGray := rl.Color{0x20, 0x2e, 0x37, 0xff}
		//rl.Scalef(spriteScale, spriteScale, 1)
		//text := "01:49"
		//fontSize := 10.0 * 3
		//textSize := rl.MeasureTextEx(countdownFont, text, fontSize, 1.0)
		//rl.DrawTextPro(
		//  countdownFont,
		//  "01:49",
		//  Vec2{-0.5 * textSize.X, -textSize.Y},
		//  Vec2{0, 0},
		//  0,
		//  fontSize,
		//  1.0,
		//  darkGray,
		//)

		rl.PopMatrix()
	}

	// Resources
	SortComponent(func(a *Resource, b *Resource) bool {
		return a.DrawOrder < b.DrawOrder
	})
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

		tex := playerTexture
		if player.Flying {
			tex = playerShipTexture
		}
		texWidth := float64(tex.Width)
		texHeight := float64(tex.Height)
		texSource := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  texWidth,
			Height: texHeight,
		}
		if player.Transmitting {
			newTexWidth := player.TransmitWidthScale * texWidth
			texSource.X = 0.5 * (texWidth - newTexWidth)
			texSource.Width = newTexWidth
			texWidth = newTexWidth
		}
		destWidth := texWidth * spriteScale
		destHeight := texHeight * spriteScale
		if player.Transmitting {
			destHeight /= Pow(player.TransmitWidthScale, 3)
		}
		texDest := rl.Rectangle{
			X:      -0.5 * destWidth,
			Y:      0.5*playerSize.Y - destHeight,
			Width:  destWidth,
			Height: destHeight,
		}
		if !player.Flying && player.FlipH {
			texSource.Width = -texSource.Width
		}
		color := rl.White
		if player.Transmitting {
			color = rl.Color{0x4f, 0x8f, 0xff, 0xff}
			rl.BeginBlendMode(rl.BLEND_ADDITIVE)
		}
		rl.DrawTexturePro(tex, texSource, texDest, Vec2{0, 0}, 0, color)
		if player.Transmitting {
			rl.EndBlendMode()
		}

		rl.PopMatrix()
	})

	// Planet terrains
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
				rl.Color4ub(planet.InnerColor.R, planet.InnerColor.G, planet.InnerColor.B, planet.InnerColor.A)
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
				rl.DrawTexturePro(bitTex, bitTexSource, bitTexDest, bitTexOrigin, bit.Rot, planet.BitsColor)
			}
		}

		rl.PopMatrix()
	})

	// Player HUD
	Each(func(ent Entity, player *Player, playerLay *Layout) {
		if player.Transmitting {
			return
		}

		// Inventory icons
		{
			screenSize := Vec2{408.0, 230.0}

			rl.PushMatrix()
			worldCameraTopLeft := rl.GetScreenToWorld2D(Vec2{0, 0}, gameCamera)
			rl.Translatef(worldCameraTopLeft.X, worldCameraTopLeft.Y, 0)
			rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
			rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

			mousePos := rl.GetMousePosition().
				Divide(Vec2{float64(rl.GetScreenWidth()), float64(rl.GetScreenHeight())}).
				Multiply(screenSize)

			iconSize := float64(elementTypes[0].iconTexture.Width)
			fontSize := 0.5 * iconSize
			iconScreenMargin := 0.5 * iconSize
			iconPos := Vec2{iconScreenMargin, iconScreenMargin}
			for typeId, amount := range player.ElementAmounts {
				elementType := &elementTypes[typeId]
				tex := elementType.iconTexture

				rl.DrawTextureEx(tex, iconPos, 0, 1, rl.White)
				rl.DrawTextureEx(elementFrameTexture, iconPos.SubtractValue(1), 0, 1, rl.White)

				textPos := iconPos.Add(Vec2{0, 1.25 * iconSize})
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

				if mousePos.X >= iconPos.X && mousePos.Y >= iconPos.Y &&
					mousePos.X <= iconPos.X+iconSize && mousePos.Y <= iconPos.Y+iconSize {
					globalHintMessages = []GlobalHintMessage{
						{Message: elementType.Name},
					}
				}

				iconPos.X += 1.375 * iconSize
			}

			rl.PopMatrix()
		}

		// Supernova countdown
		{
			rl.PushMatrix()
			worldCameraTopCenter := rl.GetScreenToWorld2D(Vec2{0.5 * float64(rl.GetScreenWidth()), 0}, gameCamera)
			rl.Translatef(worldCameraTopCenter.X, worldCameraTopCenter.Y, 0)
			rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
			rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

			iconSize := float64(elementTypes[0].iconTexture.Width)
			iconScreenMargin := 0.5 * iconSize
			rl.Translatef(0, iconScreenMargin+1.5, 0)

			totalSeconds := Max(0, Floor(player.TimeToSupernova+1))
			displaySeconds := totalSeconds % 60
			displayMinutes := totalSeconds / 60
			text := rl.TextFormat("%02d:%02d", displayMinutes, displaySeconds)

			//fontSize := iconSize
			fontSize := 10.0 * 2
			textSize := rl.MeasureTextEx(countdownFont, text, fontSize, 1.0)

			color := rl.Color{0x39, 0x4a, 0x50, 0xff}
			if totalSeconds <= 60 {
				color = rl.Color{0xa5, 0x30, 0x30, 0xff}
			}
			rl.DrawTextPro(
				countdownFont,
				text,
				Vec2{-0.5 * textSize.X, 0},
				Vec2{0, 0},
				0,
				fontSize,
				1.0,
				color,
			)

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
						mouseFramePos.X <= iconPos.X+iconSize && mouseFramePos.Y <= iconPos.Y+iconSize {
						if rl.IsMouseButtonPressed(rl.MOUSE_BUTTON_LEFT) {
							player.BuildUISelectedTypeId = ResourceTypeId(typeId)
						}
						resourceType := &resourceTypes[typeId]
						globalHintMessages = []GlobalHintMessage{}
						globalHintMessages = append(globalHintMessages, GlobalHintMessage{
							Message: resourceType.Name,
						})
						for _, amount := range resourceType.ElementAmounts {
							color := rl.Color{0x81, 0x97, 0x96, 0xff}
							if amount.Amount > player.ElementAmounts[amount.TypeId] {
								color = rl.Color{0xa5, 0x30, 0x30, 0xff}
							}
							globalHintMessages = append(globalHintMessages, GlobalHintMessage{
								Message: string(rl.TextFormat("%d %s", amount.Amount, cString(elementTypes[amount.TypeId].Name))),
								Color:   color,
							})
						}
					}

					iconPos.X += iconSize + iconSpacing
				}
			}

			if rl.IsMouseButtonPressed(rl.MOUSE_BUTTON_LEFT) {
				player.BuildUIEnabled = false
			}

			rl.PopMatrix()
		}

		// Arrow to ending planet
		if player.Flying {
			rl.PushMatrix()
			worldCameraBottomLeft := rl.GetScreenToWorld2D(Vec2{0, float64(rl.GetScreenHeight())}, gameCamera)
			rl.Translatef(worldCameraBottomLeft.X, worldCameraBottomLeft.Y, 0)
			rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
			rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

			arrowSize := float64(arrowTexture.Width)
			rl.Translatef(arrowSize, -arrowSize, 0)

			targetPos := Vec2{0, 0}
			Each(func(ent Entity, arrowTarget *ArrowTarget, targetLay *Layout) {
				targetPos = targetLay.Pos
			})
			targetDelta := targetPos.Subtract(playerLay.Pos)
			targetAngle := Atan2(targetDelta.X, -targetDelta.Y)
			rl.Rotatef((targetAngle-player.CameraRot)*180/Pi, 0, 0, 1)

			rl.DrawTextureEx(arrowTexture, Vec2{-0.5 * arrowSize, -0.5 * arrowSize}, 0, 1, rl.White)

			rl.PopMatrix()
		}
	})

	// Interaction hint
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

	// Global hint
	if len(globalHintMessages) != 0 {
		rl.PushMatrix()
		worldCameraTopRight := rl.GetScreenToWorld2D(Vec2{float64(rl.GetScreenWidth()), 0}, gameCamera)
		rl.Translatef(worldCameraTopRight.X, worldCameraTopRight.Y, 0)
		rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
		rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

		iconSize := float64(elementTypes[0].iconTexture.Width) // Same as inventory HUD
		fontSize := 0.5 * iconSize
		iconScreenMargin := 0.5 * iconSize
		rl.Translatef(-iconScreenMargin, iconScreenMargin, 0)

		textY := 0.0
		for _, message := range globalHintMessages {
			textSize := rl.MeasureTextEx(rl.GetFontDefault(), message.Message, fontSize, 1.0)
			rl.DrawTextPro(
				rl.GetFontDefault(),
				message.Message,
				Vec2{-textSize.X, textY},
				Vec2{0, 0},
				0,
				fontSize,
				1.0,
				message.Color,
			)
			textY += textSize.Y
		}

		rl.PopMatrix()

		globalHintMessages = []GlobalHintMessage{}
	}

	// Main menu
	if menuActive {
		screenSize := Vec2{408.0, 230.0}

		rl.PushMatrix()
		worldCameraTopLeft := rl.GetScreenToWorld2D(Vec2{0, 0}, gameCamera)
		rl.Translatef(worldCameraTopLeft.X, worldCameraTopLeft.Y, 0)
		rl.Rotatef(-gameCamera.Rotation, 0, 0, 1)
		rl.Scalef(2*gameCameraZoom*spriteScale, 2*gameCameraZoom*spriteScale, 1)

		mousePos := rl.GetMousePosition().
			Divide(Vec2{float64(rl.GetScreenWidth()), float64(rl.GetScreenHeight())}).
			Multiply(screenSize)

		// Title
		{
			texWidth := float64(titleTexture.Width)
			texHeight := float64(titleTexture.Height)

			texSource := rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  texWidth,
				Height: texHeight,
			}
			texDest := rl.Rectangle{
				Width:  texWidth,
				Height: texHeight,
			}
			texDest.X = 0.5 * (screenSize.X - texDest.Width)
			texDest.Y = 0.5*(screenSize.Y-texDest.Height) - 1.2*texDest.Height
			rl.DrawTexturePro(titleTexture, texSource, texDest, Vec2{0, 0}, 0, rl.White)
		}

		// Play button
		{
			texWidth := float64(playButtonTexture.Width)
			texHeight := float64(playButtonTexture.Height)

			texSource := rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  texWidth,
				Height: texHeight / 3,
			}
			texDest := rl.Rectangle{
				Width:  texWidth,
				Height: texHeight / 3,
			}
			texDest.X = 0.5 * (screenSize.X - texDest.Width)
			texDest.Y = 0.5*(screenSize.Y-texDest.Height) + 0.1*texHeight
			if mousePos.X >= texDest.X && mousePos.Y >= texDest.Y &&
				mousePos.X <= texDest.X+texDest.Width && mousePos.Y <= texDest.Y+texDest.Height {
				if rl.IsMouseButtonDown(rl.MOUSE_BUTTON_LEFT) {
					texSource.Y = texSource.Height
				} else {
					texSource.Y = 2 * texSource.Height
				}
				if rl.IsMouseButtonReleased(rl.MOUSE_BUTTON_LEFT) {
					menuActive = false
					restartGameplay = true
				}
			}
			rl.DrawTexturePro(playButtonTexture, texSource, texDest, Vec2{0, 0}, 0, rl.White)
		}

		rl.PopMatrix()
	}

	// Reticle
	{
		reticleScale := gameCameraZoom * spriteScale
		reticlePos := rl.GetScreenToWorld2D(rl.GetMousePosition(), gameCamera)
		reticleWidth := float64(reticleTexture.Width) * reticleScale
		reticleHeight := float64(reticleTexture.Height) * reticleScale
		reticleTopLeft := reticlePos.Subtract(Vec2{reticleWidth, reticleHeight}.Scale(0.5))
		rl.DrawTextureEx(reticleTexture, reticleTopLeft, 0, reticleScale, rl.Color{0x81, 0x97, 0x96, 0xff})
	}

	// End camera
	rl.EndMode2D()

	// End bloom if enabled
	if bloomEnabled {
		rl.EndTextureMode()
		source := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float64(screenTexture.Texture.Width),
			Height: -float64(screenTexture.Texture.Height),
		}
		rl.BeginShaderMode(bloomShader)
		rl.SetShaderValue(bloomShader, rl.GetShaderLocation(bloomShader, "brightness"),
			&bloomBrightness, rl.SHADER_ATTRIB_FLOAT)
		rl.DrawTextureRec(screenTexture.Texture, source, Vec2{0, 0}, rl.Red)
		rl.EndShaderMode()
	}
}
