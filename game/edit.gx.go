package game

import (
	"github.com/nikki93/dream-hotel/core/edit"
	. "github.com/nikki93/dream-hotel/core/entity"
	"github.com/nikki93/dream-hotel/core/rl"
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
	Each(func(ent Entity, move *edit.Move, circle *Circle) {
		circle.Pos = circle.Pos.Add(move.Delta)
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
	Each(func(ent Entity, circle *Circle) {
		edit.MergeBox(ent, rl.Rectangle{
			X:      circle.Pos.X - circle.Radius,
			Y:      circle.Pos.Y - circle.Radius,
			Width:  2 * circle.Radius,
			Height: 2 * circle.Radius,
		})
	})
}

//
// Draw
//

func drawGameEdit() {
}
