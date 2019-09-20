package matrix

import (
	"testing"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func TestMatrix(t *testing.T) {
	minos, err := mino.Generate(4)
	if err != nil {
		t.Errorf("failed to generate minos: %s", err)
	}

	m := NewMatrix(10, 20, 20)

	piece := mino.NewPiece(&minos[0], &mino.Point{0, 0})

	err = m.Add(piece, mino.BlockSolidBlue, mino.Point{3, 3}, false)
	if err != nil {
		t.Errorf("failed to add initial mino to matrix: %s", err)
	}

	err = m.Add(piece, mino.BlockSolidBlue, mino.Point{3, 3}, false)
	if err == nil {
		t.Error("failed to detect collision when adding second mino to matrix")
	}

	err = m.Add(piece, mino.BlockSolidBlue, mino.Point{9, 9}, false)
	if err == nil {
		t.Error("failed to detect out of bounds when adding third mino to matrix")
	}
}
