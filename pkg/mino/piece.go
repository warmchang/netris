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
	PieceJ
	PieceL
	PieceS
	PieceT
	PieceZ

	PieceJLSTZ
)

var AllRotationPivots = map[PieceType][]Point{
	PieceI: {{1, 0}, {0, 2}, {2, 0}, {0, 1}},
	PieceO: {{0, 0}, {0, 1}, {1, 1}, {2, 0}},
	PieceJ: {{1, 0}, {0, 1}, {1, 1}, {1, 1}},
	PieceL: {{1, 0}, {0, 1}, {1, 1}, {1, 1}},
	PieceS: {{1, 0}, {0, 1}, {1, 1}, {1, 1}},
	PieceT: {{1, 0}, {0, 1}, {1, 1}, {1, 1}},
	PieceZ: {{1, 0}, {0, 1}, {1, 1}, {1, 1}},
}

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

	Color int

	Rotation int
	Pivots   []Point
	Offsets  []RotationOffsets

	sync.Mutex
}

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m *Mino, loc *Point) *Piece {
	p := &Piece{Mino: m, Point: loc, Color: 0}

	offsetType := PieceJLSTZ
	pieceType := PieceT
	switch m.String() {
	case TetrominoI:
		offsetType = PieceI
		pieceType = PieceI
	case TetrominoO:
		offsetType = PieceO
		pieceType = PieceO
	case TetrominoJ:
		pieceType = PieceJ
	case TetrominoL:
		pieceType = PieceL
	case TetrominoS:
		pieceType = PieceS
	case TetrominoZ:
		pieceType = PieceZ
	}

	p.Pivots = AllRotationPivots[pieceType]
	p.Offsets = AllRotationOffsets[offsetType]

	return p
}

func (p *Piece) Rotate(deg int) *Mino {
	p.Lock()
	defer p.Unlock()

	if deg == 0 {
		return p.Mino
	}

	rotations := 1
	if deg == 270 { // TODO: Implement reverse formula
		rotations = 3
	} else if deg == 180 {
		rotations = 2
	}

	pp := p.Pivots[p.Rotation]
	px, py := pp.X, pp.Y

	w, h := p.Mino.Size()
	maxSize := w
	if h > maxSize {
		maxSize = h
	}

	newMino := make(Mino, len(*p.Mino))
	copy(newMino, *p.Mino)

	for i := 0; i < len(*p.Mino); i++ {
		for j := 0; j < rotations; j++ {
			newMino[i] = Point{newMino[i].Y + px - py, px + py - newMino[i].X + py - maxSize}
		}
	}

	return &newMino
}

func (p *Piece) ApplyRotation(deg int) {

	if deg == 0 {
		return
	}

	rotations := 1
	if deg == 270 { // TODO: Implement reverse formula
		rotations = 3
	} else if deg == 180 {
		rotations = 2
	}

	p.Rotation = (p.Rotation + rotations) % RotationStates
}
