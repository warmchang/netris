package game

import (
	"math/rand"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"git.sr.ht/~tslocum/netris/pkg/player"
)

type Game struct {
	Rank     int
	Minos    []mino.Mino
	Seed     int64
	Players  []*player.Player
	Scores   []int
	Previews []*mino.Matrix
	Matrixes []*mino.Matrix
	Bags     []*mino.Bag
	FallTime time.Duration

	Event chan interface{}

	tickers []*time.Ticker
	*sync.RWMutex
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
		Rank:    rank,
		Minos:   minos,
		Seed:    seed,
		Event:   make(chan interface{}, 10),
		RWMutex: new(sync.RWMutex)}

	g.FallTime = 850 * time.Millisecond

	g.Bags = append(g.Bags, bag)

	g.Scores = append(g.Scores, 0)

	g.Previews = append(g.Previews, mino.NewMatrix(rank, rank, 0, 1, g.Bags, g.Event, true))
	g.Matrixes = append(g.Matrixes, mino.NewMatrix(10, 20, 20, 1, g.Bags, g.Event, false))

	return g, nil
}

func (g *Game) Start() {
	g.Lock()
	defer g.Unlock()

	go g.handle(0)
}

func (g *Game) handle(player int) {
	var (
		ticker *time.Ticker
		moved  = g.Matrixes[player].Moved[player]
	)
	ticker = time.NewTicker(g.FallTime)
	for {
		select {
		case <-moved:
			ticker.Stop()
			ticker = time.NewTicker(g.FallTime)
			continue
		case <-ticker.C:
		}

		g.Matrixes[player].LowerPiece(player)
	}
}
