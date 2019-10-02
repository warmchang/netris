package game

import (
	"log"
	"net"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/player"
)

type Server struct {
	I []player.ServerInterface

	In  chan<- player.GameCommand
	Out <-chan player.GameCommand

	NewPlayers chan net.Conn

	Games map[int]*Game

	listeners []net.Listener
	sync.RWMutex
}

func NewServer(si []player.ServerInterface) *Server {
	in := make(chan player.GameCommand, player.CommandQueueSize)
	out := make(chan player.GameCommand, player.CommandQueueSize)

	s := &Server{I: si, In: in, Out: out}

	s.NewPlayers = make(chan net.Conn, player.CommandQueueSize)

	go s.accept()

	for _, serverInterface := range si {
		serverInterface.Host(s.NewPlayers)
	}

	return s
}

func (s *Server) NewGame() (*Game, error) {
	var gameID int
	for {
		if _, ok := s.Games[gameID]; !ok {
			break
		}

		gameID++
	}

	seed := time.Now().UTC().UnixNano()

	g, err := NewGame(4, seed)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (s *Server) AvailableGame() *Game {
	for _, g := range s.Games {
		if g == nil {
			continue
		}

		return g
	}

	return nil
}

func (s *Server) accept() {
	for {
		conn := <-s.NewPlayers

		p := player.NewPlayer(conn)

		g := s.AvailableGame()
		if g == nil {
			var err error
			g, err = s.NewGame()
			if err != nil {
				panic(err)
			}
		}

		g.AddPlayer(p)

		log.Println("New player", p)
		_ = p
	}
}

func (s *Server) ListenUnix(path string) {
	unixListener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	s.listeners = append(s.listeners, unixListener)

	for {
		conn, err := unixListener.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}

		s.NewPlayers <- conn
	}
}

func (s *Server) StopListening() {
	for i := range s.listeners {
		s.listeners[i].Close()
	}
}
