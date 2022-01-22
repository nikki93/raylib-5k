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

type Planet struct {
	Behavior

	Radius float64
}

type Player struct {
	Behavior
}
