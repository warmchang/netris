package mino

import (
	"fmt"
	"sync"
	"time"
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
	PieceUnknown PieceType = iota
	PieceI
	PieceO
	PieceJ
	PieceL
	PieceS
	PieceT
	PieceZ
)

var AllRotationPivotsCW = [][]Point{
	PieceUnknown: {{0, 0}, {0, 0}, {0, 0}, {0, 0}},
	PieceI:       {{1, -2}, {-1, 0}, {1, -1}, {0, 0}},
	PieceO:       {{1, 0}, {1, 0}, {1, 0}, {1, 0}},
	PieceJ:       {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceL:       {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceS:       {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceT:       {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
	PieceZ:       {{1, -1}, {0, 0}, {1, 0}, {1, 0}},
}

var AllRotationPivotsCCW = [][]Point{
	PieceUnknown: {{0, 0}, {0, 0}, {0, 0}, {0, 0}},
	PieceI:       {{2, 1}, {-1, 00}, {2, 2}, {1, 3}},
	PieceO:       {{0, 1}, {0, 1}, {0, 1}, {0, 1}},
	PieceJ:       {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceL:       {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceS:       {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
	PieceT:       {{1, 1}, {0, 2}, {1, 2}, {1, 2}},
	PieceZ:       {{1, 1}, {0, 0}, {1, 2}, {1, 2}},
}

// Rotation offsets
var AllOffsets = []Point{{0, 0}, {-1, 0}, {1, 0}, {0, -1}, {-1, -1}, {1, -1}, {-2, 0}, {2, 0}}

type Piece struct {
	Point    `json:"pp,omitempty"`
	Mino     `json:"pm,omitempty"`
	Ghost    Block `json:"pg,omitempty"`
	Solid    Block `json:"ps,omitempty"`
	Rotation int   `json:"pr,omitempty"`

	original  Mino
	pivotsCW  []Point
	pivotsCCW []Point
	resets    int
	lastReset time.Time
	landing   bool
	landed    bool

	sync.Mutex `json:"-"`
}

type LockedPiece *Piece

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m Mino, loc Point) *Piece {
	p := &Piece{Mino: m, original: m, Point: loc}

	var pieceType PieceType
	switch m.Canonical().String() {
	case TetrominoI:
		pieceType = PieceI
		p.Solid = BlockSolidCyan
		p.Ghost = BlockGhostCyan
	case TetrominoO:
		pieceType = PieceO
		p.Solid = BlockSolidYellow
		p.Ghost = BlockGhostYellow
	case TetrominoJ:
		pieceType = PieceJ
		p.Solid = BlockSolidBlue
		p.Ghost = BlockGhostBlue
	case TetrominoL:
		pieceType = PieceL
		p.Solid = BlockSolidOrange
		p.Ghost = BlockGhostOrange
	case TetrominoS:
		pieceType = PieceS
		p.Solid = BlockSolidGreen
		p.Ghost = BlockGhostGreen
	case TetrominoT:
		pieceType = PieceT
		p.Solid = BlockSolidMagenta
		p.Ghost = BlockGhostMagenta
	case TetrominoZ:
		pieceType = PieceZ
		p.Solid = BlockSolidRed
		p.Ghost = BlockGhostRed
	default:
		p.Solid = BlockSolidYellow
		p.Ghost = BlockGhostYellow
	}

	p.pivotsCW = AllRotationPivotsCW[pieceType]
	p.pivotsCCW = AllRotationPivotsCCW[pieceType]

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

	var rotationMatrix Point
	if direction == 0 {
		rotationMatrix = Point{1, -1}
	} else {
		rotationMatrix = Point{-1, 1}
	}

	var (
		rotationPivot = p.Rotation
		pivotPoint    Point
		x, y          int
	)
	for j := 0; j < rotations; j++ {
		if direction == 0 {
			rotationPivot += j
		} else {
			rotationPivot -= j
		}
		if rotationPivot < 0 {
			rotationPivot += RotationStates
		}

		if (rotationPivot == 3 && direction == 0) || (rotationPivot == 1 && direction == 1) {
			newMino = p.original
		} else {
			if direction == 0 {
				pivotPoint = p.pivotsCW[rotationPivot%RotationStates]
			} else {
				pivotPoint = p.pivotsCCW[rotationPivot%RotationStates]
			}

			for i := 0; i < len(newMino); i++ {
				x = newMino[i].X - pivotPoint.X
				y = newMino[i].Y - pivotPoint.Y

				newMino[i] = Point{y * rotationMatrix.X, x * rotationMatrix.Y}
			}
		}
	}

	return newMino
}

func (p *Piece) ApplyReset() {
	p.Lock()
	defer p.Unlock()

	if !p.landing || p.resets >= 15 {
		return
	}

	p.resets++
	p.lastReset = time.Now()
}

func (p *Piece) ApplyRotation(rotations int, direction int) {
	p.Lock()
	defer p.Unlock()

	if direction == 1 {
		rotations *= -1
	}

	p.Rotation = p.Rotation + rotations
	if p.Rotation < 0 {
		p.Rotation += RotationStates
	}
	p.Rotation %= RotationStates
}

func (p *Piece) SetLocation(x int, y int) {
	p.Lock()
	defer p.Unlock()

	p.X = x
	p.Y = y
}
