package mino

import "fmt"

type Point struct {
	X, Y int
}

func (p Point) rotate90() Point  { return Point{p.Y, -p.X} }
func (p Point) rotate180() Point { return Point{-p.X, -p.Y} }
func (p Point) rotate270() Point { return Point{-p.Y, p.X} }
func (p Point) reflect() Point   { return Point{-p.X, p.Y} }

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// Neighborhood returns the Von Neumann neighborhood of a point
func (p Point) Neighborhood() Mino {
	return Mino{
		{p.X - 1, p.Y},
		{p.X + 1, p.Y},
		{p.X, p.Y - 1},
		{p.X, p.Y + 1}}
}
