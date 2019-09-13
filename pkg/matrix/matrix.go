package matrix

import (
	"fmt"
	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func (m *Matrix) I(x int, y int) int {
	return (x * m.W) + y
}

type Matrix struct {
	W, H, B int // TODO: Implement buffer zone

	M map[int]mino.Block
}

func NewMatrix(w int, h int, b int) *Matrix {
	return &Matrix{W: w, H: h, B: b, M: make(map[int]mino.Block)}
}

func (m *Matrix) Print() {
for x := 0; x < m.W; x++ {
	for y := 0; y < m.W; y++ {
		fmt.Print(m.M[m.I(x, y)])
	}
	fmt.Println()
}
}