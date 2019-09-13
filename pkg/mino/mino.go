package mino

import (
	"sort"
	"strings"
)

type Mino []Point

func (m Mino) String() string {
	var s strings.Builder
	for i, p := range m {
		if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(p.String())
	}

	return s.String()
}

func (m Mino) Len() int      { return len(m) }
func (m Mino) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Mino) Less(i, j int) bool {
	return m[i].Y < m[j].Y || (m[i].Y == m[j].Y && m[i].X < m[j].X)
}

func (m Mino) Width() int {
	w := 0
	for _, p := range m {
		if p.X > w {
			w = p.X
		}
	}

	return w
}

func (m Mino) Render() string {
	sort.Sort(m)

	var b strings.Builder
	b.WriteString(" ")

	c := Point{0, 0}
	for _, p := range m {
		if p.Y > c.Y {
			b.WriteString("\n ")
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
	newMino := make(Mino, len(m))
	for i, p := range m {
		newMino[i] = Point{p.X - minx, p.Y - miny}
	}
	sort.Sort(newMino)
	return newMino
}
func (m Mino) variations() []Mino {
	rr := make([]Mino, 8)
	for i := 0; i < 8; i++ {
		rr[i] = make(Mino, len(m))
	}
	copy(rr[0], m)
	for j := 0; j < len(m); j++ {
		rr[1][j] = m[j].rotate90()
		rr[2][j] = m[j].rotate180()
		rr[3][j] = m[j].rotate270()
		rr[4][j] = m[j].reflect()
		rr[5][j] = m[j].rotate90().reflect()
		rr[6][j] = m[j].rotate180().reflect()
		rr[7][j] = m[j].rotate270().reflect()
	}
	return rr
}

func (m Mino) canonical() Mino {
	rr := m.variations()
	minr := rr[0].translateToOrigin()
	mins := minr.String()
	for i := 1; i < 8; i++ {
		r := rr[i].translateToOrigin()
		s := r.String()
		if s < mins {
			minr = r
			mins = s
		}
	}
	return minr
}

func (m Mino) newPoints() Mino {
	newMino := Mino{}
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
	pts := m.newPoints()
	res := make([]Mino, len(pts))
	for i, pt := range pts {
		poly := make(Mino, len(m))
		copy(poly, m)
		poly = append(poly, pt)
		res[i] = poly.canonical()
	}
	return res
}
