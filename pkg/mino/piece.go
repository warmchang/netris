package mino

import (
	"fmt"
	"sync"
)

const (
	Rotation0 = 0
	RotationR = 1
	Rotation2 = 2
	RotationL = 3

	RotationStates = 4
)

type RotationOffsets []Point

type PieceType int

const (
	PieceI PieceType = iota
	PieceO
	PieceJLSTZ
)

var AllRotationOffsets = map[PieceType][]RotationOffsets{
	PieceI: {
		{{0, 0}, {-1, 0}, {-1, 1}, {0, 1}},
		{{-1, 0}, {0, 0}, {1, 1}, {0, 1}},
		{{2, 0}, {0, 0}, {-2, 1}, {0, 1}},
		{{-1, 0}, {0, 1}, {1, 0}, {0, -1}},
		{{2, 0}, {0, -2}, {-2, 0}, {0, 2}}},
	PieceO: {{{0, 0}, {0, -1}, {-1, -1}, {-1, 0}}},
	PieceJLSTZ: {
		{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		{{0, 0}, {1, 0}, {0, 0}, {-1, 0}},
		{{0, 0}, {1, -1}, {0, 0}, {-1, -1}},
		{{0, 0}, {0, 2}, {0, 0}, {0, 2}},
		{{0, 0}, {1, 2}, {0, 0}, {-1, 2}}}}

type Piece struct {
	*Point
	*Mino

	Pivot *Point
	Color int

	Rotation int
	Offsets  RotationOffsets

	sync.Mutex
}

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m *Mino, loc *Point) *Piece {
	return &Piece{Mino: m, Point: loc, Color: 0, Pivot: &Point{1, 1}}
}

func (p *Piece) Rotate(deg int) *Mino {
	p.Lock()
	defer p.Unlock()

	if deg == 0 {
		return p.Mino
	}

	pp := p.Pivot
	px, py := pp.X, pp.Y

	w, h := p.Mino.Size()
	maxSize := w
	if h > maxSize {
		maxSize = h
	}

	rotations := 1
	if deg == 270 { // TODO: Implement reverse formula
		rotations = 3
	} else if deg == 180 {
		rotations = 2
	}

	newMino := make(Mino, len(*p.Mino))
	copy(newMino, *p.Mino)

	for i := 0; i < len(*p.Mino); i++ {
		for j := 0; j < rotations; j++ {
			newMino[i] = Point{newMino[i].Y + px - py, px + py - newMino[i].X + py - maxSize}
		}
	}

	p.Rotation = (p.Rotation + rotations) % RotationStates

	return &newMino
}
