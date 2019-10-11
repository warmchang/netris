package mino

import (
	"testing"
)

func TestMatrix(t *testing.T) {
	m, err := NewTestMatrix()
	if err != nil {
		t.Error(err)
	}

	err = m.Add(m.P[0], BlockSolidBlue, Point{3, 3}, false)
	if err != nil {
		t.Errorf("failed to add initial mino to matrix: %s", err)
	}

	err = m.Add(m.P[0], BlockSolidBlue, Point{3, 3}, false)
	if err == nil {
		t.Error("failed to detect collision when adding second mino to matrix")
	}

	err = m.Add(m.P[0], BlockSolidBlue, Point{9, 9}, false)
	if err == nil {
		t.Error("failed to detect out of bounds when adding third mino to matrix")
	}

	m.Clear()

	for i := 0; i < 8; i++ {
		ok := m.RotatePiece(0, 1, 0)
		if !ok {
			t.Errorf("failed to rotate piece on iteration %d", i)
		}
	}

	// TODO: Add rotate and wall kick tests
}
