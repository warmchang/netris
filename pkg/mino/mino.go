package mino

import (
	"sort"
	"strconv"
	"strings"
)

type Mino []Point

const (
	Monomino = "(0,0)"

	Domino = "(0,0),(1,0)"

	TrominoI = "(0,0),(1,0),(2,0)"
	TrominoL = "(0,0),(0,1),(1,1)"

	TetrominoI = "(0,0),(1,0),(2,0),(3,0)"
	TetrominoO = "(0,0),(1,0),(0,1),(1,1)"
	TetrominoT = "(0,0),(1,0),(2,0),(1,1)"
	TetrominoS = "(0,0),(1,0),(1,1),(2,1)"
	TetrominoZ = "(1,0),(2,0),(0,1),(1,1)"
	TetrominoJ = "(0,0),(1,0),(2,0),(0,1)"
	TetrominoL = "(0,0),(1,0),(2,0),(2,1)"

	PentominoF = "(1,0),(1,1),(2,1),(0,2),(1,2)"
	PentominoE = "(1,0),(0,1),(1,1),(1,2),(2,2)"
	PentominoJ = "(0,0),(0,1),(1,1),(2,1),(3,1)"
	PentominoL = "(3,0),(0,1),(1,1),(2,1),(3,1)"
	PentominoI = "(0,0),(1,0),(2,0),(3,0),(4,0)"
	PentominoN = "(0,0),(1,0),(1,1),(2,1),(3,1)"
	PentominoG = "(2,0),(3,0),(0,1),(1,1),(2,1)"
	PentominoP = "(0,0),(1,0),(0,1),(1,1),(2,1)"
	PentominoB = "(1,0),(2,0),(0,1),(1,1),(2,1)"
	PentominoS = "(1,0),(2,0),(1,1),(0,2),(1,2)"
	PentominoT = "(1,0),(1,1),(0,2),(1,2),(2,2)"
	PentominoU = "(0,0),(2,0),(0,1),(1,1),(2,1)"
	PentominoV = "(0,0),(0,1),(0,2),(1,2),(2,2)"
	PentominoW = "(0,0),(0,1),(1,1),(1,2),(2,2)"
	PentominoX = "(1,0),(0,1),(1,1),(2,1),(1,2)"
	PentominoY = "(2,0),(0,1),(1,1),(2,1),(3,1)"
	PentominoR = "(1,0),(0,1),(1,1),(2,1),(3,1)"
	PentominoZ = "(0,0),(1,0),(1,1),(1,2),(2,2)"
)

func (m Mino) Equal(other Mino) bool {
	if len(m) != len(other) {
		return false
	}

	for i := 0; i < len(m); i++ {
		if !m.HasPoint(other[i]) {
			return false
		}
	}

	return true
}

func (m Mino) String() string {
	sort.Sort(m)

	var b strings.Builder
	for i, p := range m.translateToOrigin() {
		if i > 0 {
			b.WriteRune(',')
		}

		b.WriteRune('(')
		b.WriteString(strconv.Itoa(p.X))
		b.WriteRune(',')
		b.WriteString(strconv.Itoa(p.Y))
		b.WriteRune(')')
	}

	return b.String()
}

func (m Mino) Len() int      { return len(m) }
func (m Mino) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Mino) Less(i, j int) bool {
	return m[i].Y < m[j].Y || (m[i].Y == m[j].Y && m[i].X < m[j].X)
}

func (m Mino) Size() (int, int) {
	var x, y int
	for _, p := range m {
		if p.X > x {
			x = p.X
		}
		if p.Y > y {
			y = p.Y
		}
	}

	return x + 1, y + 1
}

func (m Mino) Render() string {
	sort.Sort(m)

	var b strings.Builder
	b.WriteRune(' ')

	w, h := m.Size()

	c := Point{0, h - 1}
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			if y < c.Y {
				b.WriteRune('\n')
				b.WriteRune(' ')

				c.X = 0
			}
			if x > c.X {
				for i := c.X; i < x; i++ {
					b.WriteRune(' ')
				}
			}

			c.X = x + 1
			c.Y = y
			b.WriteRune('X')
		}
	}

	return b.String()
}

func (m Mino) HasPoint(p Point) bool {
	for _, mp := range m {
		if mp == p {
			return true
		}
	}

	return false
}

func (m Mino) minCoords() (int, int) {
	minx := m[0].X
	miny := m[0].Y
	for i := 1; i < len(m); i++ {
		if m[i].X < minx {
			minx = m[i].X
		}
		if m[i].Y < miny {
			miny = m[i].Y
		}
	}
	return minx, miny
}

func (m Mino) translateToOrigin() Mino {
	minx, miny := m.minCoords()
	for i, p := range m {
		m[i].X = p.X - minx
		m[i].Y = p.Y - miny
	}
	return m
}

func (m Mino) Rotate(deg int) Mino {
	var rotateFunc func(Point) Point
	switch deg {
	case 90:
		rotateFunc = Point.Rotate90
	case 180:
		rotateFunc = Point.Rotate180
	case 270:
		rotateFunc = Point.Rotate270
	default:
		return m
	}

	for i := 0; i < len(m); i++ {
		m[i] = rotateFunc(m[i])
	}

	return m
}

func (m Mino) Variations() []Mino {
	v := make([]Mino, 3)
	for i := 0; i < 3; i++ {
		v[i] = make(Mino, len(m))
	}

	for j := 0; j < len(m); j++ {
		v[0][j] = m[j].Rotate90()
		v[1][j] = m[j].Rotate180()
		v[2][j] = m[j].Rotate270()
	}

	return v
}

func (m Mino) Canonical() Mino {
	var (
		ms = m.String()
		c  = -1
		v  = m.Variations()
		vs string
	)

	for i := 0; i < 3; i++ {
		vs = v[i].String()
		if vs < ms {
			c = i
			ms = vs
		}
	}

	if c == -1 {
		return m.Flatten()
	}

	return v[c].Flatten()
}

func (m Mino) Flatten() Mino {
	w, h := m.Size()

	var top, right, bottom, left int
	for i := 0; i < len(m); i++ {
		if m[i].Y == 0 {
			bottom++
		} else if m[i].Y == (h - 1) {
			top++
		}

		if m[i].X == 0 {
			left++
		} else if m[i].X == (w - 1) {
			right++
		}
	}

	flattest := bottom
	var rotate int
	if left > flattest {
		flattest = left
		rotate = 270
	}
	if top > flattest {
		flattest = top
		rotate = 180
	}
	if right > flattest {
		flattest = right
		rotate = 90
	}
	if rotate > 0 {
		m = m.Rotate(rotate)
	}

	return m
}

func (m Mino) newPoints() Mino {
	var newMino Mino

	for _, p := range m {
		n := p.Neighborhood()
		for _, np := range n {
			if !m.HasPoint(np) {
				newMino = append(newMino, np)
			}
		}
	}

	return newMino
}

func (m Mino) newMinos() []Mino {
	mino := make(Mino, len(m))
	copy(mino, m)

	points := m.newPoints()
	minos := make([]Mino, len(points))

	for i, p := range points {
		minos[i] = append(mino, p).Canonical()
	}

	return minos
}
