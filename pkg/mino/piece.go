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

var AllRotationPivotsCW = map[PieceType][]Point{
	PieceI: {{1, -2}, {-1, 0}, {1, -1}, {0, 0}},
	PieceO: {{1, 0}, {1, 0}, {1, 0}, {1, 0}},
	PieceJ: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceL: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceS: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceT: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceZ: {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
}

// IN PROGRESS
var AllRotationPivotsCCW = map[PieceType][]Point{
	PieceI: {{2, 1}, {-1, 00}, {2, 2}, {1, 3}},
	PieceO: {{0, 1}, {0, 1}, {0, 1}, {0, 1}},
	PieceJ: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceL: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceS: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceT: {{1, 1}, {0, 2}, {1, 2}, {1, 2}},
	PieceZ: {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
}

// AllRotationOffets is a list of all piece offsets.  Each set includes offsets
// for 0, R, L and 2 rotation states.
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
	Point
	Mino
	Original Mino

	Rotation  int
	PivotsCW  []Point
	PivotsCCW []Point
	Offsets   []RotationOffsets

	Color int

	sync.Mutex
}

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m Mino, loc Point) *Piece {
	p := &Piece{Mino: m, Original: m, Point: loc, Color: 0}

	offsetType := PieceJLSTZ
	var pieceType PieceType
	switch m.Canonical().String() {
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
	case TetrominoT:
		pieceType = PieceT
	case TetrominoZ:
		pieceType = PieceZ
	}

	p.PivotsCW = AllRotationPivotsCW[pieceType]
	p.PivotsCCW = AllRotationPivotsCCW[pieceType]
	p.Offsets = AllRotationOffsets[offsetType]

	return p
}

// Rotate returns the new mino of a piece when a rotation is applied
func (p *Piece) Rotate(rotations int, direction int) Mino {
	p.Lock()
	defer p.Unlock()

	if rotations == 0 {
		return p.Mino
	}

	newMino := make(Mino, len(p.Mino))
	copy(newMino, p.Mino.Origin())

	w, h := newMino.Size()
	maxSize := w
	if h > maxSize {
		maxSize = h
	}

	var rotationPivot int
	for j := 0; j < rotations; j++ {
		if direction == 0 {
			rotationPivot = p.Rotation + j
		} else {
			rotationPivot = p.Rotation - j
		}

		if rotationPivot < 0 {
			rotationPivot += RotationStates
		}

		if (rotationPivot == 3 && direction == 0) || (rotationPivot == 1 && direction == 1) {
			newMino = p.Original
		} else {
			pp := p.PivotsCW[rotationPivot%RotationStates]
			if direction == 1 {
				pp = p.PivotsCCW[rotationPivot%RotationStates]
			}
			px, py := pp.X, pp.Y

			for i := 0; i < len(newMino); i++ {
				x := newMino[i].X
				y := newMino[i].Y

				if direction == 0 {
					newMino[i] = Point{(0 * (x - px)) + (1 * (y - py)), (-1 * (x - px)) + (0 * (y - py))}
				} else {
					newMino[i] = Point{(0 * (x - px)) + (-1 * (y - py)), (1 * (x - px)) + (0 * (y - py))}
				}
			}
		}
	}

	return newMino
}

func (p *Piece) ApplyRotation(rotations int, direction int) {
	if direction == 1 {
		rotations *= -1
	}

	p.Rotation = p.Rotation + rotations
	if p.Rotation < 0 {
		p.Rotation += RotationStates
	}
	p.Rotation %= RotationStates
}
