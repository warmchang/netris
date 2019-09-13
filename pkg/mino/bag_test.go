package mino

import (
	"testing"
)

func TestBag(t *testing.T) {
	var (
		minos []Mino
		err   error
	)
	for _, d := range minoTestData {
		minos, err = Generate(d.Rank)
		if err != nil {
			t.Errorf("failed to generate minos: %s", err)
		}

		if len(minos) != d.Minos {
			t.Error("failed to generate minos: unexpected number of minos generated")
		}
		minos, err := Generate(d.Rank)
		if err != nil {
			t.Errorf("failed to create minos for bag: %s", err)
		}

		b := NewBag(minos)
		taken := make(map[string]int)
		for i := 1; i < 4; i++ {
			for i := 0; i < d.Minos; i++ {
				mino := b.Take()
				taken[mino.String()]++
			}

			if len(taken) != d.Minos {
				t.Errorf("minos placed in bag do not match minos taken - placed: %s - taken: %v", b.Minos, taken)
			}

			for _, mino := range minos {
				if taken[mino.String()] != i {
					t.Fatalf("minos placed in bag do not match minos taken - placed: %s - taken: %v", minos, taken)
				}
			}
		}
	}
}
