package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"code.rocketnine.space/tslocum/netris/pkg/event"
)

const (
	DefaultPort   = 1984
	DefaultServer = "playnetris.com"
)

type Server struct {
	I []ServerInterface

	In  chan GameCommandInterface
	Out chan GameCommandInterface

	Games map[int]*Game

	Logger chan string

	listeners  []net.Listener
	NewPlayers chan *IncomingPlayer

	created time.Time

	logLevel int

	sync.RWMutex
}

type IncomingPlayer struct {
	Name string
	Conn *Conn
}

type ServerInterface interface {
	// Load config
	Host(newPlayers chan<- *IncomingPlayer)
	Shutdown(reason string)
}

func NewServer(si []ServerInterface, logLevel int) *Server {
	in := make(chan GameCommandInterface, CommandQueueSize)
	out := make(chan GameCommandInterface, CommandQueueSize)

	s := &Server{I: si, In: in, Out: out, Games: make(map[int]*Game), created: time.Now(), logLevel: logLevel}

	var (
		g   *Game
		err error
	)
	g, err = s.NewGame()
	if err != nil {
		log.Fatal(err)
	}

	g.Eternal = true
	g.Name = "No speed limit"

	g, err = s.NewGame()
	if err != nil {
		log.Fatal(err)
	}

	g.Eternal = true
	g.Name = "Speed limit 100"
	g.SpeedLimit = 100

	g, err = s.NewGame()
	if err != nil {
		log.Fatal(err)
	}

	g.Eternal = true
	g.Name = "Speed limit 40"
	g.SpeedLimit = 40

	s.NewPlayers = make(chan *IncomingPlayer, CommandQueueSize)

	go s.accept()
	go s.handle()
	for _, serverInterface := range si {
		serverInterface.Host(s.NewPlayers)
	}

	return s
}

func (s *Server) NewGame() (*Game, error) {
	s.Lock()
	defer s.Unlock()

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

	g.ID = gameID
	g.LogLevel = s.logLevel

	s.Games[gameID] = g

	return g, nil
}

func (s *Server) handle() {
	for {
		time.Sleep(1 * time.Minute)

		s.removeTerminatedGames()
	}
}

func (s *Server) removeTerminatedGames() {
	s.Lock()
	defer s.Unlock()

	for gameID, g := range s.Games {
		g.Lock()
		if !g.Terminated {
			g.Unlock()
			continue
		}

		delete(s.Games, gameID)
		g.Unlock()
	}
}

func (s *Server) FindGame(p *Player, gameID int, newGame ListedGame) *Game {
	var (
		g   *Game
		err error
	)

	if newGame.Name != "" {
		// Create a custom game
		g, err = s.NewGame()
		if err != nil {
			log.Fatalf("failed to create custom game: %s", err)
		}

		g.Lock()

		g.Name = GameName(newGame.Name)

		g.MaxPlayers = newGame.MaxPlayers
		if g.MaxPlayers < 0 {
			g.MaxPlayers = 0
		} else if g.MaxPlayers > 999 {
			g.MaxPlayers = 999
		}

		g.SpeedLimit = newGame.SpeedLimit
		if g.SpeedLimit < 0 {
			g.SpeedLimit = 0
		} else if g.SpeedLimit > 999 {
			g.SpeedLimit = 999
		}

		g.Unlock()
	} else if gameID > 0 {
		// Join a game by its ID
		s.Lock()
		gm := s.Games[gameID]
		s.Unlock()

		if gm != nil {
			gm.Lock()
			canJoin := !gm.Terminated && (gm.MaxPlayers == 0 || len(gm.Players) < gm.MaxPlayers)
			gm.Unlock()

			if canJoin {
				g = gm
			} else {
				p.Write(&GameCommandMessage{Message: "Failed to join game - Player limit reached"})
				return nil
			}
		} else {
			p.Write(&GameCommandMessage{Message: "Failed to join game - Invalid game ID"})
			return nil
		}
	} else if gameID == 0 {
		// Join any game
		s.Lock()
		for _, gm := range s.Games {
			gm.Lock()
			if !gm.Terminated && (gm.MaxPlayers == 0 || len(gm.Players) < gm.MaxPlayers) {
				gm.Unlock()
				g = gm
				break
			}

			gm.Unlock()
		}
		s.Unlock()
	} else {
		// Create a local game
		g, err = s.NewGame()
		if err != nil {
			log.Fatalf("failed to create local game: %s", err)
		}

		g.Local = true
	}

	if g == nil {
		p.Write(&GameCommandMessage{Message: "Failed to join game"})
		return nil
	}

	g.Lock()

	g.AddPlayerL(p)

	if gameID == event.GameIDNewLocal {
		go g.Start(0)
	} else if len(g.Players) > 1 {
		go s.initiateAutoStart(g)
	} else if !g.Started {
		p.Write(&GameCommandMessage{Message: "Waiting for at least two players to join..."})
	}

	g.Unlock()

	return g
}

func (s *Server) accept() {
	for {
		np := <-s.NewPlayers

		p := NewPlayer(np.Name, np.Conn)

		go s.handleNewPlayer(p)
	}
}

func (s *Server) handleNewPlayer(pl *Player) {
	handled := false
	go func() {
		time.Sleep(10 * time.Second)
		if !handled {
			pl.Close()
		}
	}()

	for e := range pl.In {
		switch e.Command() {
		case CommandListGames:
			if _, ok := e.(*GameCommandListGames); ok {
				var gl []*ListedGame

				s.Lock()
				for _, g := range s.Games {
					g.Lock()
					if g.Terminated {
						g.Unlock()
						continue
					}

					gl = append(gl, &ListedGame{ID: g.ID, Name: g.Name, Players: len(g.Players), MaxPlayers: g.MaxPlayers, SpeedLimit: g.SpeedLimit})
					g.Unlock()
				}
				s.Unlock()

				sort.Slice(gl, func(i, j int) bool {
					if gl[i].Players == gl[j].Players {
						return gl[i].Name < gl[j].Name
					}

					return gl[i].Players > gl[j].Players
				})

				pl.Write(&GameCommandListGames{Games: gl})
			}
		case CommandJoinGame:
			if p, ok := e.(*GameCommandJoinGame); ok {
				pl.Name = Nickname(p.Name)

				g := s.FindGame(pl, p.GameID, p.Listing)
				if g == nil {
					return
				}

				if p.Listing.Name == "" {
					g.Logf(LogStandard, "Player %s joined %s", pl.Name, g.Name)
				} else {
					g.Logf(LogStandard, "Player %s created new game %s", pl.Name, g.Name)
				}

				go s.handleGameCommands(pl, g)

				handled = true
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
	var (
		msgJSON []byte
		err     error
	)
	for e := range pl.In {
		c := e.Command()
		if (c != CommandPing && c != CommandPong && c != CommandUpdateMatrix) || g.LogLevel >= LogVerbose {
			msgJSON, err = json.Marshal(e)
			if err != nil {
				log.Fatal(err)
			}

			g.Logf(LogStandard, "%d -> %s %s", e.Source(), e.Command(), msgJSON)
		}

		g.Lock()

		switch p := e.(type) {
		case *GameCommandDisconnect:
			g.RemovePlayerL(p.SourcePlayer)
		case *GameCommandMessage:
			if player, ok := g.Players[p.SourcePlayer]; ok {
				s.Logf("<%s> %s", player.Name, p.Message)

				msg := strings.ReplaceAll(strings.TrimSpace(p.Message), "\n", "")
				if msg != "" {
					g.WriteAllL(&GameCommandMessage{Player: p.SourcePlayer, Message: msg})
				}
			}
		case *GameCommandNickname:
			if player, ok := g.Players[p.SourcePlayer]; ok {
				newNick := Nickname(p.Nickname)
				if newNick != "" && newNick != player.Name {
					oldNick := player.Name
					player.Name = newNick

					g.Logf(LogStandard, "* %s is now known as %s", oldNick, newNick)
					g.WriteAllL(&GameCommandNickname{Player: p.SourcePlayer, Nickname: newNick})
				}
			}
		case *GameCommandUpdateMatrix:
			if pl, ok := g.Players[p.SourcePlayer]; ok {
				for _, m := range p.Matrixes {
					pl.Matrix.Replace(m)

					if g.SpeedLimit > 0 && m.Speed > g.SpeedLimit+5 && time.Since(g.TimeStarted) > 7*time.Second {
						pl.Matrix.SetGameOver()

						g.WriteMessage(fmt.Sprintf("%s went too fast and crashed", pl.Name))
						g.WriteAllL(&GameCommandGameOver{Player: p.SourcePlayer})
					}
				}

				m := pl.Matrix
				spawn := m.SpawnLocation(m.P)
				if m.P != nil && spawn.X >= 0 && spawn.Y >= 0 && m.P.X != spawn.X {
					pl.Moved = time.Now()
					pl.Idle = 0
				}
			}
		case *GameCommandGameOver:
			g.Players[p.SourcePlayer].Matrix.SetGameOver()

			g.WriteMessage(fmt.Sprintf("%s was knocked out", g.Players[p.SourcePlayer].Name))
			g.WriteAllL(&GameCommandGameOver{Player: p.SourcePlayer})
		case *GameCommandSendGarbage:
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
		case *GameCommandStats:
			go func(p *Player) {
				players := 0
				games := 0

				s.Lock()
				for _, g := range s.Games {
					players += len(g.Players)
					games++
				}
				s.Unlock()

				p.Write(&GameCommandStats{Created: s.created, Players: players, Games: games})
			}(g.Players[p.SourcePlayer])
		}

		g.Unlock()
	}
}

func (s *Server) Listen(address string) {
	var network string
	network, address = NetworkAndAddress(address)

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", address, err)
	}

	s.listeners = append(s.listeners, listener)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
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

func NetworkAndAddress(address string) (string, string) {
	var network string
	if strings.ContainsAny(address, `\/`) {
		network = "unix"
	} else {
		network = "tcp"

		if !strings.Contains(address, `:`) {
			address = fmt.Sprintf("%s:%d", address, DefaultPort)
		}
	}

	return network, address
}
