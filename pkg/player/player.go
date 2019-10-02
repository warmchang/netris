package player

import (
	"bufio"
	"log"
	"net"
)

const CommandQueueSize = 10

type Player struct {
	Conn net.Conn

	Client ClientInterface

	In  chan<- GameCommand
	Out <-chan GameCommand

	Name string
}

type ConnectingPlayer struct {
	Client ClientInterface

	Name string
}

func NewPlayer(conn net.Conn) *Player {
	in := make(chan GameCommand, CommandQueueSize)
	out := make(chan GameCommand, CommandQueueSize)

	p := &Player{Conn: conn, In: in, Out: out}

	go p.handleRead()
	go p.handleWrite()

	return p
}

func (p *Player) handleRead() {
	scanner := bufio.NewScanner(p.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		log.Println("read ", line)
	}
}

func (p *Player) handleWrite() {
	for e := range p.Out {
		_ = e
		log.Println("player write", e)
		p.Conn.Write([]byte("test\n"))
	}
}

type ClientInterface interface {
	Attach(in chan<- GameCommand, out <-chan GameCommand)
	Detach(reason string)
}

type ServerInterface interface {
	// Load config
	Host(newPlayers chan<- net.Conn)
	Shutdown(reason string)
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
