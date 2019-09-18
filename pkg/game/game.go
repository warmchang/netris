package game

import (
	"math/rand"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/mino"

	"git.sr.ht/~tslocum/netris/pkg/player"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
)

type Game struct {
	Rank       int
	Minos      []mino.Mino
	Seed       int64
	Players    map[int]*player.Player
	Previews   map[int]*matrix.Matrix
	Matrixes   map[int]*matrix.Matrix
	Bags       map[int]*mino.Bag
	Pieces     map[int]*mino.Piece
	NextPieces map[int]*mino.Piece
	FallTime   time.Duration

	dropped map[int]chan bool
	tickers map[int]*time.Ticker
	sync.RWMutex
}

func NewGame(rank int, seed int64) (*Game, error) {
	rand.Seed(seed)

	minos, err := mino.Generate(rank)
	if err != nil {
		return nil, err
	}

	bag, err := mino.NewBag(minos)
	if err != nil {
		return nil, err
	}

	g := &Game{
		Rank:       rank,
		Minos:      minos,
		Seed:       seed,
		Players:    make(map[int]*player.Player),
		Previews:   make(map[int]*matrix.Matrix),
		Matrixes:   make(map[int]*matrix.Matrix),
		Bags:       make(map[int]*mino.Bag),
		Pieces:     make(map[int]*mino.Piece),
		NextPieces: make(map[int]*mino.Piece),
		tickers:    make(map[int]*time.Ticker),
		dropped:    make(map[int]chan bool)}

	g.FallTime = 850 * time.Millisecond

	g.Bags[0] = bag

	g.Previews[0] = matrix.NewMatrix(rank, rank, 0)
	g.Matrixes[0] = matrix.NewMatrix(10, 20, 20)

	return g, nil
}

func (g *Game) Start() {
	g.Lock()
	defer g.Unlock()

	g.takePiece(0)

	g.dropped[0] = make(chan bool, 2)

	go g.handle(0)
}

func (g *Game) DroppedPiece(player int) {
	g.dropped[player] <- true
}

func (g *Game) handle(player int) {
	var (
		ticker *time.Ticker
	)
	ticker = time.NewTicker(g.FallTime)
	for {
		select {
		case <-g.dropped[player]:
			ticker.Stop()
			ticker = time.NewTicker(g.FallTime)
			continue
		case <-ticker.C:
		}

		g.lowerPiece(player)
	}
}

func (g *Game) lowerPiece(player int) {
	if g.Matrixes[0].CanAdd(g.Pieces[0], mino.Point{g.Pieces[0].X, g.Pieces[0].Y - 1}) {
		g.Pieces[0].Y -= 1
	} else {
		g.landPiece(player)
	}
}

func (g *Game) landPiece(player int) {
	solidBlock := g.Pieces[0].SolidBlock()

	dropped := false
	for y := 0; y < g.Matrixes[0].H+g.Matrixes[0].B && y <= g.Pieces[0].Y; y++ {
		if g.Matrixes[0].CanAdd(g.Pieces[0], mino.Point{g.Pieces[0].X, y}) {
			err := g.Matrixes[0].Add(g.Pieces[0], solidBlock, mino.Point{g.Pieces[0].X, y}, false)
			if err != nil {
				panic(err)
			}

			dropped = true
			break
		}
	}

	if !dropped {
		panic("failed to land piece")
		return
	}

	g.dropped[player] <- true

	g.takePiece(player)

	cleared := g.Matrixes[0].ClearFilled()
	if cleared > 0 {
		// TODO Send cleared event
	}
}

func (g *Game) LandPiece(player int) {
	g.Lock()
	defer g.Unlock()

	g.landPiece(player)
}

func (g *Game) takePiece(player int) {
	g.Pieces[player] = mino.NewPiece(g.Bags[player].Take(), mino.Point{4, 17})
	g.NextPieces[player] = mino.NewPiece(g.Bags[player].Next(), mino.Point{0, 0})
}

func (g *Game) TakePiece(player int) {
	g.Lock()
	defer g.Unlock()

	g.takePiece(player)
}
