package mino

import "fmt"

type Piece struct {
	Point
	Mino

	Color int
}

func (p *Piece) String() string {
	return fmt.Sprintf("%+v", *p)
}

func NewPiece(m Mino, loc Point) *Piece {
	return &Piece{Mino: m, Point: loc, Color: 0}
}

func (p *Piece) Rotate(deg int) {
	// TODO: Rotate around pivot point and translate to orgin, adjusting loc as necessary, return bool if rotation is possible
	p.Mino = p.Mino.Rotate(deg)
}
