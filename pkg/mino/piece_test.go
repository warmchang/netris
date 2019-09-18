package mino

import (
	"testing"
)

func TestPiece(t *testing.T) {
	minos, err := Generate(4)
	if err != nil {
		t.Errorf("failed to generate minos for rank %d: %s", 4, err)
	}

	if len(minos) != 7 {
		t.Errorf("failed to generate minos for rank %d: unexpected number of minos generated", 4)
	}

	if minos[2].String() != TetrominoJ {
		t.Errorf("unexpected mino found when generating J teromino: %s", minos[2])
	}

	p := NewPiece(minos[2], Point{0, 0})
	// TODO
	_ = p
}
