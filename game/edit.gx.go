package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
	. "github.com/nikki93/raylib-5k/core/geom"
	"github.com/nikki93/raylib-5k/core/rl"
)

//
// Validate
//

func validateGameEdit() {
}

//
// Moves
//

func applyGameEditMoves() {
	Each(func(ent Entity, move *edit.Move, lay *Layout) {
		lay.Pos = lay.Pos.Add(move.Delta)
	})
}

//
// Input
//

func inputGameEdit() {
}

//
// Boxes
//

func mergeGameEditBoxes() {
	// Planets
	Each(func(ent Entity, planet *Planet, lay *Layout) {
		edit.MergeBox(ent, rl.Rectangle{
			X:      lay.Pos.X - planet.Radius,
			Y:      lay.Pos.Y - planet.Radius,
			Width:  2 * planet.Radius,
			Height: 2 * planet.Radius,
		})
	})

	// Resources
	Each(func(ent Entity, resource *Resource, lay *Layout) {
		resourceType := &resourceTypes[resource.TypeId]
		texture := resourceType.texture
		texSize := Vec2{float64(texture.Width), float64(texture.Height)}
		destSize := texSize.Scale(spriteScale)
		edit.MergeBox(ent, rl.Rectangle{
			X:      lay.Pos.X - 0.5*destSize.X,
			Y:      lay.Pos.Y - destSize.Y,
			Width:  destSize.X,
			Height: destSize.Y,
		})
	})

	// Player
	Each(func(ent Entity, player *Player, lay *Layout) {
		edit.MergeBox(ent, rl.Rectangle{
			X:      lay.Pos.X - 0.5*playerSize.X,
			Y:      lay.Pos.Y - 0.5*playerSize.Y,
			Width:  playerSize.X,
			Height: playerSize.Y,
		})
	})
}

//
// Draw
//

func drawGameEdit() {
}
