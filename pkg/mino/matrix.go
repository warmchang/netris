package mino

import (
	"fmt"
	"strings"
	"sync"

	"git.sr.ht/~tslocum/netris/pkg/event"
)

func I(x int, y int, w int) int {
	return (y * w) + x
}

type Matrix struct {
	W, H, B int // TODO: Implement buffer zone

	M map[int]Block // Matrix
	O map[int]Block // Overlay

	Bags []*Bag
	P    []*Piece // Pieces

	Moved []chan int

	preview bool

	Event chan<- interface{}

	*sync.RWMutex
}

func NewMatrix(w int, h int, b int, players int, bags []*Bag, event chan<- interface{}, preview bool) *Matrix {
	m := Matrix{W: w, H: h, B: b, M: make(map[int]Block), Bags: bags, Event: event, preview: preview, RWMutex: new(sync.RWMutex)}

	m.P = make([]*Piece, players)

	m.Moved = make([]chan int, players)
	for i := 0; i < players; i++ {
		m.Moved[i] = make(chan int, 10)
	}

	m.takePiece(0)

	return &m
}

func (m *Matrix) takePiece(player int) {
	if m.preview {
		return
	}

	p := NewPiece(m.Bags[player].Take(), Point{0, m.H - 1})
	p.X = m.PieceStartX(p)

	m.P[player] = p
}

func (m *Matrix) TakePiece(player int) {
	m.Lock()
	defer m.Unlock()

	m.takePiece(player)
}

func (m *Matrix) CanAddAt(mn *Piece, loc Point) bool {
	m.Lock()
	defer m.Unlock()

	return m.canAddAt(mn, loc)
}

func (m *Matrix) canAddAt(mn *Piece, loc Point) bool {
	if loc.Y < 0 {
		return false
	}

	var (
		x, y  int
		index int
	)

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return false
		}

		index = I(x, y, m.W)
		if m.M[index] != BlockNone {
			return false
		}
	}

	return true
}

func (m *Matrix) CanAdd(mn *Piece) bool {
	m.Lock()
	defer m.Unlock()

	var (
		x, y  int
		index int
	)

	for _, p := range mn.Mino {
		x = p.X + mn.X
		y = p.Y + mn.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return false
		}

		index = I(x, y, m.W)
		if m.M[index] != BlockNone {
			return false
		}
	}

	return true
}

func (m *Matrix) Add(mn *Piece, b Block, loc Point, overlay bool) error {
	m.Lock()
	defer m.Unlock()

	return m.add(mn, b, loc, overlay)
}

func (m *Matrix) add(mn *Piece, b Block, loc Point, overlay bool) error {
	var (
		x, y  int
		index int

		newM map[int]Block
	)

	if overlay {
		newM = m.NewO()
	} else {
		newM = m.NewM()
	}

	for _, p := range mn.Mino {
		x = p.X + loc.X
		y = p.Y + loc.Y

		if x < 0 || x >= m.W || y < 0 || y >= m.H+m.B {
			return fmt.Errorf("failed to add to matrix at %s: point %s out of bounds (%d, %d)", loc, p, x, y)
		}

		index = I(x, y, m.W)
		if !overlay && newM[index] != BlockNone {
			return fmt.Errorf("failed to add to matrix at %s: point %s already contains %s", loc, p, newM[index])
		}

		newM[index] = b
	}

	if overlay {
		m.O = newM
	} else {
		m.M = newM
	}

	return nil
}

func (m *Matrix) Empty(loc Point) bool {
	index := I(loc.X, loc.Y, m.W)
	return m.M[index] == BlockNone
}

func (m *Matrix) LineFilled(y int) bool {
	for x := 0; x < m.W; x++ {
		if m.Empty(Point{x, y}) {
			return false
		}
	}

	return true
}

func (m *Matrix) ClearFilled() int {
	m.Lock()
	defer m.Unlock()

	return m.clearFilled()
}

func (m *Matrix) clearFilled() int {
	cleared := 0

	for y := 0; y < m.H+m.B; y++ {
		for {
			if m.LineFilled(y) {
				for my := y + 1; my < m.H+m.B; my++ {
					for mx := 0; mx < m.W; mx++ {
						m.M[I(mx, my-1, m.W)] = m.M[I(mx, my, m.W)]
					}
				}

				cleared++
				continue
			}

			break
		}
	}

	return cleared
}

func (m *Matrix) ClearOverlay() {
	m.Lock()
	defer m.Unlock()

	for i := range m.O {
		m.O[i] = BlockNone
	}
}

func (m *Matrix) Clear() {
	m.Lock()
	defer m.Unlock()

	for i := range m.M {
		m.M[i] = BlockNone
	}
}

func (m *Matrix) NewM() map[int]Block {
	newM := make(map[int]Block, len(m.M))
	for i, b := range m.M {
		newM[i] = b
	}

	return newM
}

func (m *Matrix) NewO() map[int]Block {
	newO := make(map[int]Block, len(m.O))
	for i, b := range m.O {
		newO[i] = b
	}

	return newO
}

func (m *Matrix) Block(x int, y int) Block {
	index := I(x, y, m.W)

	if b, ok := m.O[index]; ok && b != BlockNone {
		return b
	}

	return m.M[index]
}

func (m *Matrix) Rotate(player int, rotations int, direction int) bool {
	if rotations == 0 {
		return false
	}

	p := m.P[player]

	originalMino := make(Mino, len(p.Mino))
	copy(originalMino, p.Mino)

	var rotationOffset int
	if direction == 0 {
		rotationOffset = (p.Rotation + rotations)
	} else {
		rotationOffset = (p.Rotation - rotations)
		if rotationOffset < 0 {
			rotationOffset += RotationStates
		}
	}
	rotationOffset %= RotationStates
	if rotationOffset >= len(p.Offsets) {
		rotationOffset = 0
	}

	p.Mino = p.Rotate(rotations, direction)

	if m.canAddAt(p, p.Point) {
		p.ApplyRotation(rotations, direction)

		return true
	}

	var offX, offY int
	for i := 0; i < len(p.Offsets); i++ {
		if direction == 0 {
			offX = p.Offsets[i][p.Rotation].X - p.Offsets[i][rotationOffset].X
			offY = p.Offsets[i][p.Rotation].Y - p.Offsets[i][rotationOffset].Y
		} else {
			offX = p.Offsets[i][rotationOffset].X - p.Offsets[i][p.Rotation].X
			offY = p.Offsets[i][rotationOffset].Y - p.Offsets[i][p.Rotation].Y
		}

		px := p.X + offX
		py := p.Y + offY

		if m.canAddAt(p, Point{px, py}) {
			p.X = px
			p.Y = py

			p.ApplyRotation(rotations, direction)

			return true
		}

	}

	p.Mino = originalMino
	return false
}

func (m *Matrix) PieceStartX(p *Piece) int {
	w, _ := p.Size()
	return (m.W / 2) - (w / 2)

}

func (m *Matrix) Render() string {
	m.RLock()
	defer m.RUnlock()

	var b strings.Builder

	for y := m.H - 1; y >= 0; y-- {
		for x := 0; x < m.W; x++ {
			b.WriteRune(m.Block(x, y).Rune())
		}

		if y == 0 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}

// LowerPiece lowers the active piece by one line when possible, otherwise the
// piece is landed
func (m *Matrix) LowerPiece(player int) {
	m.Lock()
	defer m.Unlock()

	if m.canAddAt(m.P[0], Point{m.P[0].X, m.P[0].Y - 1}) {
		m.movePiece(0, 0, -1)
	} else {
		m.landPiece(player)
	}
}

func (m *Matrix) landPiece(player int) {
	solidBlock := m.P[0].SolidBlock()

	dropped := false
LANDPIECE:
	for y := m.P[0].Y; y >= 0; y-- {
		if y == 0 || !m.canAddAt(m.P[0], Point{m.P[0].X, y - 1}) {
			for dropY := y - 1; dropY < m.H+m.B; dropY++ {
				if !m.canAddAt(m.P[0], Point{m.P[0].X, dropY}) {
					continue
				}

				err := m.add(m.P[0], solidBlock, Point{m.P[0].X, dropY}, false)
				if err != nil {
					panic(err)
				}

				dropped = true
				break LANDPIECE
			}
		}
	}

	if !dropped {
		panic("failed to land piece")
		return
	}

	//m.takePiece(player)

	cleared := m.clearFilled()

	score := 0
	switch cleared {
	case 0:
		// No score
	case 1:
		score = 100
	case 2:
		score = 300
	case 3:
		score = 500
	case 4:
		score = 800
	default:
		score = 1000 + ((cleared - 5) * 200)
	}

	m.Moved[player] <- score
	if cleared > 0 {
		ev := event.ScoreEvent{Score: score}
		ev.Player = 0

		m.Event <- &ev
	}

	m.takePiece(player)
}

func (m *Matrix) MovePiece(player int, x int, y int) bool {
	m.Lock()
	defer m.Unlock()

	return m.movePiece(player, x, y)
}

func (m *Matrix) movePiece(player int, x int, y int) bool {
	px := m.P[player].X + x
	py := m.P[player].Y + y

	if !m.canAddAt(m.P[player], Point{px, py}) {
		return false
	}

	m.P[0].X = px
	m.P[0].Y = py

	if y < 0 {
		m.Moved[player] <- 0
	}

	return true
}

func (m *Matrix) LandPiece(player int) {
	m.Lock()
	defer m.Unlock()

	m.landPiece(player)
}
