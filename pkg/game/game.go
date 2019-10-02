package game

import (
	"fmt"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"

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

	Buffer []string

	Event chan interface{}

	tickers []*time.Ticker
	*sync.RWMutex
}

func NewGame(rank int, seed int64) (*Game, error) {
	minos, err := mino.Generate(rank)
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

	return g, nil
}

func (g *Game) AddPlayer(p *player.Player) {
	g.Lock()
	defer g.Unlock()

	bag, err := mino.NewBag(g.Seed, g.Minos)
	if err != nil {
		return
	}

	g.Players = append(g.Players, p)

	g.Bags = append(g.Bags, bag)

	g.Scores = append(g.Scores, 0)

	g.Previews = append(g.Previews, mino.NewMatrix(g.Rank, g.Rank, 0, 1, g.Bags, g.Event, true))
	g.Matrixes = append(g.Matrixes, mino.NewMatrix(10, 20, 20, 1, g.Bags, g.Event, false))

}

func (g *Game) Start() {
	g.Lock()
	defer g.Unlock()

	go g.handle()

	go g.handleLowerPiece(0)
}

func (g *Game) handle() {
	var e interface{}
	for {
		e = <-g.Event
		if ev, ok := e.(*event.ScoreEvent); ok {
			g.Scores[ev.Player] += ev.Score

			if ev.Message != "" {
				g.Buffer = append(g.Buffer, ev.Message)
			}
		} else if ev, ok := e.(*event.Event); ok {
			if ev.Message != "" {
				g.Buffer = append(g.Buffer, ev.Message)
			}
		} else {
			panic(fmt.Sprintf("unknown event type: %+v", e))
		}
	}
}

func (g *Game) handleLowerPiece(player int) {
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
