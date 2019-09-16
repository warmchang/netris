package game

import (
	"math/rand"

	"git.sr.ht/~tslocum/netris/pkg/player"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
)

type Game struct {
	Seed     int64
	Players  map[int]player.Player
	Matrixes map[int]matrix.Matrix
}

func NewGame(seed int64) *Game {
	g := &Game{Seed: seed, Players: make(map[int]player.Player), Matrixes: make(map[int]matrix.Matrix)}

	return g
}

func (g *Game) Start() {
	rand.Seed(g.Seed)
}
