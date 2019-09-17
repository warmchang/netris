package mino

import (
	"math/rand"
	"sync"
)

type Bag struct {
	Minos    []Mino
	Original []Mino
	sync.Mutex
}

func NewBag(minos []Mino) *Bag {
	b := &Bag{Original: minos}
	b.Shuffle()

	return b
}

func (b *Bag) Take() Mino {
	b.Lock()
	defer b.Unlock()

	mino := b.Minos[0]
	if len(b.Minos) == 1 {
		b.Shuffle()
	} else {
		b.Minos = b.Minos[1:]
	}

	return mino
}

func (b *Bag) Shuffle() {
	b.Minos = b.Original
	rand.Shuffle(len(b.Minos), func(i, j int) { b.Minos[i], b.Minos[j] = b.Minos[j], b.Minos[i] })
}