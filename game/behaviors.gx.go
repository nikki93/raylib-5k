//gx:include "core/core.hh"

package game

import (
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
)

type Circle struct {
	Behavior

	Pos    Vec2
	Radius float64
}
