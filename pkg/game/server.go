package game

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
)

type Server struct {
	I []ServerInterface

	In  chan GameCommandInterface
	Out chan GameCommandInterface

	NewPlayers chan *IncomingPlayer

	Games map[int]*Game

	Logger chan string

	listeners []net.Listener
	sync.RWMutex
}

type IncomingPlayer struct {
	Name string
	Conn *ServerConn
}

type ServerInterface interface {
	// Load config
	Host(newPlayers chan<- *IncomingPlayer)
	Shutdown(reason string)
}

func NewServer(si []ServerInterface) *Server {
	in := make(chan GameCommandInterface, CommandQueueSize)
	out := make(chan GameCommandInterface, CommandQueueSize)

	s := &Server{I: si, In: in, Out: out, Games: make(map[int]*Game)}

	s.NewPlayers = make(chan *IncomingPlayer, CommandQueueSize)

	go s.accept()

	for _, serverInterface := range si {
		serverInterface.Host(s.NewPlayers)
	}

	return s
}

func (s *Server) NewGame() (*Game, error) {
	gameID := 1
	for {
		if _, ok := s.Games[gameID]; !ok {
			break
		}

		gameID++
	}

	draw := make(chan event.DrawObject)
	go func() {
		for range draw {
		}
	}()

	logger := make(chan string, LogQueueSize)
	go func() {
		for msg := range logger {
			s.Log(fmt.Sprintf("Game %d: %s", gameID, msg))
		}
	}()

	g, err := NewGame(4, nil, logger, draw)
	if err != nil {
		return nil, err
	}

	// TODO
	//g.LogLevel = LogDebug

	s.Games[gameID] = g

	return g, nil
}

func (s *Server) FindGame(p *Player, gameID int) *Game {
	s.Lock()
	defer s.Unlock()

	var (
		g   *Game
		err error
	)

	if gm, ok := s.Games[gameID]; ok {
		g = gm
	}

	if g == nil {
		for gameID, g = range s.Games {
			if g != nil {
				if g.Terminated {
					delete(s.Games, gameID)
					g = nil

					s.Log("Cleaned up game ", gameID)
					continue
				}

				break
			}
		}
	}

	if g == nil {
		g, err = s.NewGame()
		if err != nil {
			panic(err)
		}

		if gameID == -1 {
			g.Local = true
		}
	}

	g.Lock()

	g.AddPlayerL(p)
	if len(g.Players) > 1 {
		var players []string
		for playerID, player := range g.Players {
			if playerID == p.Player {
				continue
			}

			players = append(players, player.Name)
		}

		p.Write(&GameCommandMessage{Message: "Joined game - Players: " + strings.Join(players, " ")})
	}

	if gameID == -1 {
		go g.Start(0)
	} else if len(g.Players) > 1 {
		go s.initiateAutoStart(g)
	} else if !g.Started {
		p.Write(&GameCommandMessage{Message: "Waiting for at least two players to join..."})
	}

	g.Unlock()

	return g
}

func (s *Server) joinGame(p *Player, g *Game) {
	var notified bool
	for {
		if p.Terminated {
			return
		}

		g.Lock()
		if !g.Started {
			break
		} else if !notified {
			p.Write(&GameCommandMessage{Message: "Game in progress, waiting to join next game.."})
		}

		g.Unlock()
		time.Sleep(500 * time.Millisecond)
	}

	if !g.Starting {

	}
	g.Unlock()
}

func (s *Server) accept() {
	for {
		np := <-s.NewPlayers

		p := NewPlayer(np.Name, np.Conn)

		s.Log("Incoming connection from ", np.Name)

		go s.handleJoinGame(p)
	}
}

func (s *Server) handleJoinGame(pl *Player) {
	s.Log("waiting first msg handle join game ")
	for e := range pl.In {
		s.Log("handle join game ", e.Command(), e)
		if e.Command() == CommandJoinGame {
			if p, ok := e.(*GameCommandJoinGame); ok {
				pl.Name = Nickname(p.Name)

				s.Log("JOINING GAME", p)

				g := s.FindGame(pl, p.GameID)

				s.Log("New player added to game", *pl, p.GameID)

				go s.handleGameCommands(pl, g)
				return
			}
		}
	}
}

func (s *Server) initiateAutoStart(g *Game) {
	g.Lock()
	defer g.Unlock()

	if g.Starting || g.Started {
		return
	}

	g.Starting = true

	go func() {
		g.WriteMessage("Starting game...")
		time.Sleep(2 * time.Second)
		g.Start(0)
	}()
}

func (s *Server) handleGameCommands(pl *Player, g *Game) {
	s.Log("waiting first msg handle game commands")
	for e := range pl.In {
		if e.Command() != CommandUpdateMatrix || g.LogLevel >= LogVerbose {
			s.Log("REMOTE handle game command ", e.Command(), " from ", e.Source(), e)
		}

		g.Lock()

		switch e.Command() {
		case CommandMessage:
			if p, ok := e.(*GameCommandMessage); ok {
				if player, ok := g.Players[p.SourcePlayer]; ok {
					s.Log("<" + player.Name + "> " + p.Message)

					msg := strings.ReplaceAll(strings.TrimSpace(p.Message), "\n", "")
					if msg != "" {
						g.WriteAllL(&GameCommandMessage{Player: p.SourcePlayer, Message: msg})
					}
				}
			}
		case CommandUpdateMatrix:
			if p, ok := e.(*GameCommandUpdateMatrix); ok {
				for _, m := range p.Matrixes {
					g.Players[p.SourcePlayer].Matrix.Replace(m)
				}
			}
		case CommandGameOver:
			if p, ok := e.(*GameCommandGameOver); ok {
				g.Players[p.SourcePlayer].Matrix.SetGameOver()

				g.WriteMessage(fmt.Sprintf("%s was knocked out", g.Players[p.SourcePlayer].Name))
				g.WriteAllL(&GameCommandGameOver{Player: p.SourcePlayer})
			}
		case CommandSendGarbage:
			if p, ok := e.(*GameCommandSendGarbage); ok {
				leastGarbagePlayer := -1
				leastGarbage := -1
				for playerID, player := range g.Players {
					if playerID == p.SourcePlayer || player.Matrix.GameOver {
						continue
					}

					if leastGarbage == -1 || player.totalGarbageReceived < leastGarbage {
						leastGarbagePlayer = playerID
						leastGarbage = player.totalGarbageReceived
					}
				}

				if leastGarbagePlayer != -1 {
					g.Players[leastGarbagePlayer].totalGarbageReceived += p.Lines
					g.Players[leastGarbagePlayer].pendingGarbage += p.Lines

					g.Players[p.SourcePlayer].totalGarbageSent += p.Lines
				}
			}
		}

		g.Unlock()
	}
}

func (s *Server) Listen(address string) {
	var network string
	if strings.ContainsRune(address, ':') {
		network = "tcp"
	} else {
		network = "unix"
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	s.listeners = append(s.listeners, listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}

		s.NewPlayers <- &IncomingPlayer{Name: "Anonymous", Conn: NewServerConn(conn, nil)}
	}
}

func (s *Server) StopListening() {
	for i := range s.listeners {
		s.listeners[i].Close()
	}
}

func (s *Server) Log(a ...interface{}) {
	if s.Logger == nil {
		return
	}

	s.Logger <- fmt.Sprint(a...)
}

func (s *Server) Logf(format string, a ...interface{}) {
	if s.Logger == nil {
		return
	}

	s.Logger <- fmt.Sprintf(format, a...)
}
