//gx:include "core/core.hh"

package game

import (
	. "github.com/nikki93/dream-hotel/core/geom"
	"github.com/nikki93/dream-hotel/core/rl"
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
}
