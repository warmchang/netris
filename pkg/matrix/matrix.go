package matrix

import (
	"fmt"
	"strings"
	"sync"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func I(x int, y int, w int) int {
	return (y * w) + x
}

type Matrix struct {
	W, H, B int // TODO: Implement buffer zone

	M map[int]mino.Block // Matrix
	O map[int]mino.Block // Overlay

	sync.RWMutex
}

func NewMatrix(w int, h int, b int) *Matrix {
	return &Matrix{W: w, H: h, B: b, M: make(map[int]mino.Block)}
}

func (m *Matrix) CanAdd(mn *mino.Piece, loc mino.Point) bool {
	if loc.X < 0 || loc.Y < 0 {
		return false
	}

	m.Lock()
	defer m.Unlock()

	var (
		x, y  int
		index int
	)

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x >= m.W || y >= m.H+m.B {
			return false
		}

		index = I(x, y, m.W)
		if m.M[index] != mino.BlockNone {
			return false
		}
	}

	return true
}

func (m *Matrix) Add(mn *mino.Piece, b mino.Block, loc mino.Point, overlay bool) error {
	m.Lock()
	defer m.Unlock()

	var (
		x, y  int
		index int

		newM map[int]mino.Block
	)

	if overlay {
		newM = m.NewO()
	} else {
		newM = m.NewM()
	}

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x >= m.W || y >= m.H+m.B {
			return fmt.Errorf("failed to add to matrix at %s: point %s out of bounds", loc, p)
		}

		index = I(x, y, m.W)
		if !overlay && newM[index] != mino.BlockNone {
			return fmt.Errorf("failed to add to matrix at %s: point %s already contains %s", loc, p, newM[index])
		}

		newM[index] = b
	}

	if overlay {
		m.O = newM
	} else {
		m.M = newM
	}

	return nil
}

func (m *Matrix) Empty(loc mino.Point) bool {
	index := I(loc.X, loc.Y, m.W)
	return m.M[index] == mino.BlockNone
}

func (m *Matrix) LineFilled(y int) bool {
	for x := 0; x < m.W; x++ {
		if m.Empty(mino.Point{x, y}) {
			return false
		}
	}

	return true
}

func (m *Matrix) ClearFilled() int {
	m.Lock()
	defer m.Unlock()

	cleared := 0

	for y := 0; y < m.H+m.B; y++ {
		for {
			if m.LineFilled(y) {
				for my := y + 1; my < m.H+m.B; my++ {
					for mx := 0; mx < m.W; mx++ {
						m.M[I(mx, my-1, m.W)] = m.M[I(mx, my, m.W)]
					}
				}

				cleared++
				continue
			}

			break
		}
	}

	// TODO: Empty lines

	return cleared
}

func (m *Matrix) ClearOverlay() {
	m.Lock()
	defer m.Unlock()

	for i := range m.O {
		m.O[i] = mino.BlockNone
	}
}

func (m *Matrix) ClearMatrix() {
	m.Lock()
	defer m.Unlock()

	for i := range m.M {
		m.M[i] = mino.BlockNone
	}
}

func (m *Matrix) NewM() map[int]mino.Block {
	newM := make(map[int]mino.Block, len(m.M))
	for i, b := range m.M {
		newM[i] = b
	}

	return newM
}

func (m *Matrix) NewO() map[int]mino.Block {
	newO := make(map[int]mino.Block, len(m.O))
	for i, b := range m.O {
		newO[i] = b
	}

	return newO
}

func (m *Matrix) Block(x int, y int) mino.Block {
	index := I(x, y, m.W)

	if b, ok := m.O[index]; ok && b != mino.BlockNone {
		return b
	}

	return m.M[index]
}

func (m *Matrix) Render() string {
	m.RLock()
	defer m.RUnlock()

	var b strings.Builder

	for y := m.H - 1; y >= 0; y-- {
		for x := 0; x < m.W; x++ {
			b.WriteRune(m.Block(x, y).Rune())
		}

		if y == 0 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}
