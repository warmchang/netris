package mino

import (
	"errors"
	"sync"
)

type MinoCache struct {
	m map[int][]Mino
	sync.RWMutex
}

func getCachedMinos(rank int) ([]Mino, bool) {
	cachedMinos.RLock()
	defer cachedMinos.RUnlock()

	minos, ok := cachedMinos.m[rank]
	return minos, ok
}

func resetCachedMinos() {
	cachedMinos = &MinoCache{m: make(map[int][]Mino)}
}

var cachedMinos = &MinoCache{m: make(map[int][]Mino)}

// Generate
func Generate(rank int) ([]Mino, error) {
	if minos, ok := getCachedMinos(rank); ok {
		return minos, nil
	}

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

		var minos []Mino
		found := make(map[string]bool)
		for _, mino := range r {
			for _, newMino := range mino.newMinos() {
				if s := newMino.String(); !found[s] {
					minos = append(minos, newMino.translateToOrigin())
					found[s] = true
				}
			}
		}

		cachedMinos.Lock()
		cachedMinos.m[rank] = minos
		cachedMinos.Unlock()

		return minos, nil
	}
}

func monomino() Mino {
	return Mino{{0, 0}}
}
