//gx:include "core/core.hh"

package game

import (
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
)

type Position struct {
	Behavior

	Pos Vec2
}

type Velocity struct {
	Behavior

	Vel Vec2
}

type Planet struct {
	Behavior

	Radius float64
}

type Up struct {
	Behavior

	Up Vec2
}

type Gravity struct {
	Behavior

	Strength float64
}

type Player struct {
	Behavior
}
