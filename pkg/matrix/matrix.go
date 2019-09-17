package matrix

import (
	"fmt"
	"strings"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func I(x int, y int, w int) int {
	return (y * w) + x
}

type Matrix struct {
	W, H, B int // TODO: Implement buffer zone

	M map[int]mino.Block
}

func NewMatrix(w int, h int, b int) *Matrix {
	return &Matrix{W: w, H: h, B: b, M: make(map[int]mino.Block)}
}

func (m *Matrix) Add(mn mino.Mino, b mino.Block, loc mino.Point) error {
	var (
		index int
		newM  = m.NewM()
	)
	for _, p := range mn {
		index = I(p.X, p.Y, m.W)
		if m.M[index] != mino.BlockNone {
			return fmt.Errorf("failed to add to matrix at %s: point %s already contains %s", loc, p, m.M[index])
		}

		newM[index] = b
	}

	m.M = newM

	return nil
}

func (m *Matrix) Empty(loc mino.Point) bool {
	index := I(loc.X, loc.Y, m.W)
	return m.M[index] == mino.BlockNone
}

func (m *Matrix) Clear() error {
	for i := range m.M {
		m.M[i] = mino.BlockNone
	}

	return nil
}

func (m *Matrix) NewM() map[int]mino.Block {
	newM := make(map[int]mino.Block, len(m.M))
	for i, b := range m.M {
		newM[i] = b
	}

	return newM
}

func (m *Matrix) Render() string {
	var b strings.Builder

	for y := m.B; y < (m.H + m.B); y++ {
		for x := 0; x < m.W; x++ {
			b.WriteRune(mino.BlockToRune(m.M[I(x, y, m.W)]))
		}

		if y == m.H-1 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}
