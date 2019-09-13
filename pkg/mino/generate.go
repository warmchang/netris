package mino

// Generate
func Generate(n int) []Mino {
	switch {
	case n < 0:
		panic("invalid rank")
	case n == 0:
		return []Mino{}
	case n == 1:
		return []Mino{monomino()}
	default:
		r := Generate(n - 1)
		var minos []Mino
		for _, mino := range r {
			for _, newMino := range mino.newMinos() {
				minos = append(minos, newMino.translateToOrigin())
			}
		}

		return minos
	}
}

func monomino() Mino {
	return Mino{{0, 0}}
}
