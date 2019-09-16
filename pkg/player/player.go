package player

import (
	"log"
	"sync"
)

const CommandQueueSize = 10

type Player struct {
	Client ClientInterface

	In  chan<- GameCommand
	Out <-chan GameCommand

	Name string
}

type ConnectingPlayer struct {
	Client ClientInterface

	Name string
}

func NewPlayer(c *ConnectingPlayer) *Player {
	in := make(chan GameCommand, CommandQueueSize)
	out := make(chan GameCommand, CommandQueueSize)

	p := &Player{Name: c.Name, Client: c.Client, In: in, Out: out}

	p.Client.Attach(in, out)

	return p
}

type ClientInterface interface {
	Attach(in chan<- GameCommand, out <-chan GameCommand)
	Detach(reason string)
}

type ServerInterface interface {
	// Load config
	Host(newPlayers chan<- *ConnectingPlayer)
	Shutdown(reason string)
}

type Server struct {
	I []ServerInterface

	In  chan<- GameCommand
	Out <-chan GameCommand

	sync.RWMutex
}

func NewServer(si []ServerInterface) *Server {
	in := make(chan GameCommand, CommandQueueSize)
	out := make(chan GameCommand, CommandQueueSize)

	s := &Server{I: si, In: in, Out: out}

	newPlayers := make(chan *ConnectingPlayer, CommandQueueSize)

	go s.accept(newPlayers)

	for _, serverInterface := range si {
		serverInterface.Host(newPlayers)
	}

	return s
}

func (s *Server) accept(newPlayers <-chan *ConnectingPlayer) {
	for {
		np := <-newPlayers

		p := NewPlayer(np)
		log.Println("accept", p.Name)
		log.Println(p)
	}
}

type Command int

const (
	CommandUnknown Command = 0
	CommandDisconnect
	CommandChat
	CommandNewGame
	CommandJoinGame
	CommandQuitGame
)

type GameCommand struct {
	C Command
	P interface{}
}
