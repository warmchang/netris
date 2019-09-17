package matrix

import (
	"strings"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func (m *Matrix) I(x int, y int) int {
	return (y * m.W) + x
}

type Matrix struct {
	W, H, B int // TODO: Implement buffer zone

	M map[int]mino.Block
}

func NewMatrix(w int, h int, b int) *Matrix {
	return &Matrix{W: w, H: h, B: b, M: make(map[int]mino.Block)}
}

func blockToRune(block mino.Block) rune {
	switch block {
	case mino.BlockNone:
		return ' '
	case mino.BlockGhost:
		return '▒'
	case mino.BlockSolid:
		return '█'
	default:
		return '?'
	}
}

func (m *Matrix) Render() string {
	var b strings.Builder

	for y := m.B; y < (m.H + m.B); y++ {
		for x := 0; x < m.W; x++ {
			b.WriteRune(blockToRune(m.M[m.I(x, y)]))
		}

		if y == m.H-1 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}
