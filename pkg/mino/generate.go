package mino

import (
	"errors"
)

// Generate procedurally generates minos of a supplied rank.
func Generate(rank int) ([]Mino, error) {
	switch {
	case rank < 0:
		return nil, errors.New("invalid rank")
	case rank == 0:
		return []Mino{}, nil
	case rank == 1:
		return []Mino{monomino()}, nil
	default:
		r, err := Generate(rank - 1)
		if err != nil {
			return nil, err
		}

		var (
			minos []Mino
			s     string
			found = make(map[string]bool)
		)
		for _, mino := range r {
			for _, newMino := range mino.NewMinos() {
				s = newMino.Canonical().String()
				if found[s] {
					continue
				}

				minos = append(minos, newMino.Canonical())
				found[s] = true
			}
		}

		return minos, nil
	}
}

func monomino() Mino {
	return Mino{{0, 0}}
}
