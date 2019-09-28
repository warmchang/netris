package mino

import (
	"testing"
)

func TestMatrix(t *testing.T) {
	minos, err := Generate(4)
	if err != nil {
		t.Errorf("failed to generate minos: %s", err)
	}

	bag, err := NewBag(minos)
	if err != nil {
		t.Errorf("failed to generate minos: %s", err)
	}

	ev := make(chan interface{})
	go func() {
		for range ev {
		}
	}()

	m := NewMatrix(10, 20, 20, 1, []*Bag{bag}, ev, false)

	m.P[0] = NewPiece(minos[0], Point{3, 7})

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
		ok := m.Rotate(0, 1, 0)
		if !ok {
			t.Errorf("failed to rotate piece on iteration %d", i)
		}
	}

	// TODO: Add rotate and wall kick tests
}
