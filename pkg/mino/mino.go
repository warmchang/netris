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
	TrominoL = "(0,0),(1,0),(0,1)"

	TetrominoI = "(0,0),(1,0),(2,0),(3,0)"
	TetrominoO = "(0,0),(1,0),(0,1),(1,1)"
	TetrominoT = "(0,0),(1,0),(2,0),(1,1)"
	TetrominoS = "(0,0),(1,0),(1,1),(2,1)"
	TetrominoZ = "(1,0),(2,0),(0,1),(1,1)"
	TetrominoJ = "(0,0),(1,0),(2,0),(0,1)"
	TetrominoL = "(0,0),(1,0),(2,0),(2,1)"

	PentominoF = "(0,0),(1,0),(1,1),(2,1),(1,2)"
	PentominoE = "(1,0),(2,0),(0,1),(1,1),(1,2)"
	PentominoJ = "(0,0),(1,0),(2,0),(3,0),(0,1)"
	PentominoL = "(0,0),(1,0),(2,0),(3,0),(3,1)"
	PentominoP = "(0,0),(1,0),(2,0),(0,1),(1,1)"
	PentominoZ = "(1,0),(2,0),(1,1),(0,2),(1,2)"
	PentominoI = "(0,0),(1,0),(2,0),(3,0),(4,0)"
	PentominoX = "(1,0),(0,1),(1,1),(2,1),(1,2)"
	PentominoV = "(0,0),(1,0),(2,0),(0,1),(0,2)"
	PentominoB = "(0,0),(1,0),(2,0),(1,1),(2,1)"
	PentominoN = "(1,0),(2,0),(3,0),(0,1),(1,1)"
	PentominoG = "(0,0),(1,0),(2,0),(2,1),(3,1)"
	PentominoS = "(0,0),(1,0),(1,1),(1,2),(2,2)"
	PentominoT = "(0,0),(1,0),(2,0),(1,1),(1,2)"
	PentominoU = "(0,0),(1,0),(2,0),(0,1),(2,1)"
	PentominoW = "(1,0),(2,0),(0,1),(1,1),(0,2)"
	PentominoY = "(0,0),(1,0),(2,0),(3,0),(2,1)"
	PentominoR = "(0,0),(1,0),(2,0),(3,0),(1,1)"
)

func NewMino(points string) Mino {
	var m Mino

	var last int
	for i, p := range strings.Split(strings.ReplaceAll(strings.ReplaceAll(points, "(", ""), ")", ""), ",") {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		if i%2 == 0 {
			last = v
			continue
		}

		m = append(m, Point{last, v})
	}

	return m
}

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
	b.Grow(5*len(m) + (len(m) - 1))

	for i := range m {
		if i > 0 {
			b.WriteRune(',')
		}

		b.WriteRune('(')
		b.WriteString(strconv.Itoa(m[i].X))
		b.WriteRune(',')
		b.WriteString(strconv.Itoa(m[i].Y))
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
	var (
		w, h = m.Size()
		c    = Point{0, h - 1}
		b    strings.Builder
	)
	for y := h - 1; y >= 0; y-- {
		c.X = 0
		c.Y = y

		for x := 0; x < w; x++ {
			if !m.HasPoint(Point{x, y}) {
				continue
			}

			for i := x - c.X; i > 0; i-- {
				b.WriteRune(' ')
			}

			b.WriteRune('X')
			c.X = x + 1
		}

		b.WriteRune('\n')
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

	for _, p := range m[1:] {
		if p.X < minx {
			minx = p.X
		}
		if p.Y < miny {
			miny = p.Y
		}
	}

	return minx, miny
}

func (m Mino) Origin() Mino {
	minx, miny := m.minCoords()

	newMino := make(Mino, len(m))
	for i, p := range m {
		newMino[i].X = p.X - minx
		newMino[i].Y = p.Y - miny
	}

	return newMino
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
		ms = m.Origin().String()
		c  = -1
		v  = m.Origin().Variations()
		vs string
	)

	for i := 0; i < 3; i++ {
		vs = v[i].Origin().String()
		if vs < ms {
			c = i
			ms = vs
		}
	}

	if c == -1 {
		return m.Origin().Flatten().Origin()
	}

	return v[c].Origin().Flatten().Origin()
}

func (m Mino) Flatten() Mino {
	var (
		w, h  = m.Size()
		sides [4]int // Left Top Right Bottom
	)
	for i := 0; i < len(m); i++ {
		if m[i].Y == 0 {
			sides[3]++
		} else if m[i].Y == (h - 1) {
			sides[1]++
		}

		if m[i].X == 0 {
			sides[0]++
		} else if m[i].X == (w - 1) {
			sides[2]++
		}
	}

	var (
		largestSide   = 3
		largestLength = sides[3]
	)
	for i, s := range sides[:2] {
		if s > largestLength {
			largestSide = i
			largestLength = s
		}
	}

	var rotateFunc func(Point) Point
	switch largestSide {
	case 0: // Left
		rotateFunc = Point.Rotate270
	case 1: // Top
		rotateFunc = Point.Rotate180
	case 2: // Right
		rotateFunc = Point.Rotate90
	default: // Bottom
		return m
	}

	newMino := make(Mino, len(m))
	copy(newMino, m)
	for i := 0; i < len(m); i++ {
		newMino[i] = rotateFunc(newMino[i])
	}

	return newMino
}

func (m Mino) newPoints() Mino {
	var newMino Mino

	for _, p := range m {
		for _, np := range p.Neighborhood() {
			if !m.HasPoint(np) {
				newMino = append(newMino, np)
			}
		}
	}

	return newMino
}

func (m Mino) newMinos() []Mino {
	points := m.newPoints()

	minos := make([]Mino, len(points))
	for i, p := range points {
		minos[i] = append(m, p).Canonical()
	}

	return minos
}
