package game

import (
	"github.com/nikki93/raylib-5k/core/edit"
	. "github.com/nikki93/raylib-5k/core/entity"
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
	Each(func(ent Entity, move *edit.Move, pos *Position) {
		pos.Pos = pos.Pos.Add(move.Delta)
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
	Each(func(ent Entity, planet *Planet, pos *Position) {
		edit.MergeBox(ent, rl.Rectangle{
			X:      pos.Pos.X - planet.Radius,
			Y:      pos.Pos.Y - planet.Radius,
			Width:  2 * planet.Radius,
			Height: 2 * planet.Radius,
		})
	})
}

//
// Draw
//

func drawGameEdit() {
}
