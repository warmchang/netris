package mino

import (
	"sort"
	"strconv"
	"strings"
)

type Mino []Point

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

	c := Point{0, 0}
	for _, p := range m {
		if p.Y > c.Y {
			b.WriteRune('\n')
			b.WriteRune(' ')
			c.X = 0
		}
		if p.X > c.X {
			for i := c.X; i < p.X; i++ {
				b.WriteRune(' ')
			}
		}

		c = p
		c.X++
		b.WriteRune('X')
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

func (m Mino) rotate(deg int) Mino {
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

func (m Mino) variations() []Mino {
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

func (m Mino) canonical() Mino {
	var (
		ms = m.String()
		c  = -1
		v  = m.variations()
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
		return m.flatten()
	}

	return v[c].flatten()
}

func (m Mino) flatten() Mino {
	w, h := m.Size()

	var top, right, bottom, left int
	for i := 0; i < len(m); i++ {
		if m[i].Y == 0 {
			top++
		} else if m[i].Y == (h - 1) {
			bottom++
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
		rotate = 90
	}
	if top > flattest {
		flattest = top
		rotate = 180
	}
	if right > flattest {
		flattest = right
		rotate = 270
	}
	if rotate > 0 {
		m = m.rotate(rotate)
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
		minos[i] = append(mino, p).canonical()
	}

	return minos
}
