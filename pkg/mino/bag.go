package mino

import (
	"math/rand"
	"sync"
)

type Bag struct {
	Minos    []Mino
	Original []Mino

	rand *rand.Rand
	sync.Mutex
}

func NewBag(seed int64, minos []Mino) (*Bag, error) {
	b := &Bag{Original: minos, rand: rand.New(rand.NewSource(seed))}
	b.Shuffle()

	return b, nil
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

func (b *Bag) Next() Mino {
	b.Lock()
	defer b.Unlock()

	return b.Minos[0]
}

func (b *Bag) Shuffle() {
	b.Minos = b.Original

	b.rand.Shuffle(len(b.Minos), func(i, j int) { b.Minos[i], b.Minos[j] = b.Minos[j], b.Minos[i] })
}
