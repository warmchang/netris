package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
)

const ConnTimeout = 30 * time.Second

type GameCommandTransport struct {
	Command Command `json:"cmd"`
	Data    json.RawMessage
}

type Conn struct {
	conn         net.Conn
	LastTransfer time.Time
	Terminated   bool

	Player     int
	In         chan GameCommandInterface
	out        chan GameCommandInterface
	forwardOut chan GameCommandInterface

	*sync.WaitGroup
}

func NewServerConn(conn net.Conn, forwardOut chan GameCommandInterface) *Conn {
	c := Conn{conn: conn, WaitGroup: new(sync.WaitGroup)}

	c.In = make(chan GameCommandInterface, CommandQueueSize)
	c.out = make(chan GameCommandInterface, CommandQueueSize)
	c.forwardOut = forwardOut

	c.LastTransfer = time.Now()

	if conn == nil {
		// Local instance

		go c.handleLocalWrite()
	} else {
		go c.handleRead()
		go c.handleWrite()
		go c.handleSendKeepAlive()
	}

	return &c
}

func Connect(address string) (*Conn, error) {
	var (
		network string
		conn    net.Conn
		err     error
		tries   int
	)
	network, address = NetworkAndAddress(address)

	for {
		conn, err = net.DialTimeout(network, address, ConnTimeout)
		if err != nil {
			if tries > 25 {
				return nil, fmt.Errorf("failed to connect to %s: %s", address, err)
			} else {
				time.Sleep(250 * time.Millisecond)

				tries++
				continue
			}
		}

		return NewServerConn(conn, nil), nil
	}
}

func (s *Conn) handleSendKeepAlive() {
	t := time.NewTicker(7 * time.Second)
	for {
		<-t.C

		if s.Terminated {
			t.Stop()
			return
		}

		// TODO: Only send when necessary
		s.Write(&GameCommandPing{Message: fmt.Sprintf("a%d", time.Now().UnixNano())})
	}
}

func (s *Conn) Write(gc GameCommandInterface) {
	if s == nil || s.Terminated {
		return
	}

	s.Add(1)
	s.out <- gc
}

func (s *Conn) handleLocalWrite() {
	for e := range s.out {
		if s.forwardOut != nil {
			s.forwardOut <- e
		}

		s.Done()
	}
}

func (s *Conn) addSourceID(gc GameCommandInterface) {
	gc.SetSource(s.Player)
}

func (s *Conn) handleRead() {
	if s.conn == nil {
		return
	}

	err := s.conn.SetReadDeadline(time.Now().Add(ConnTimeout))
	if err != nil {
		s.Close()
		return
	}

	var (
		msg       GameCommandTransport
		gc        GameCommandInterface
		processed bool

		um = func(mgc interface{}) {
			err := json.Unmarshal(msg.Data, mgc)
			if err != nil {
				s.Close()
			}
		}
	)
	scanner := bufio.NewScanner(s.conn)
	for scanner.Scan() {
		processed = false

		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			break
		}

		s.LastTransfer = time.Now()

		switch msg.Command {
		case CommandDisconnect:
			var mgc GameCommandDisconnect
			um(&mgc)
			gc = &mgc
		case CommandPing:
			var mgc GameCommandPing
			um(&mgc)

			s.Write(&GameCommandPong{Message: mgc.Message})
			processed = true
		case CommandPong:
			var mgc GameCommandPong
			um(&mgc)
			gc = &mgc
		case CommandMessage:
			var mgc GameCommandMessage
			um(&mgc)
			gc = &mgc
		case CommandNickname:
			var mgc GameCommandNickname
			um(&mgc)
			gc = &mgc
		case CommandJoinGame:
			var mgc GameCommandJoinGame
			um(&mgc)
			gc = &mgc
		case CommandQuitGame:
			var mgc GameCommandQuitGame
			um(&mgc)
			gc = &mgc
		case CommandUpdateGame:
			var mgc GameCommandUpdateGame
			um(&mgc)
			gc = &mgc
		case CommandStartGame:
			var mgc GameCommandStartGame
			um(&mgc)
			gc = &mgc
		case CommandGameOver:
			var mgc GameCommandGameOver
			um(&mgc)
			gc = &mgc
		case CommandUpdateMatrix:
			var mgc GameCommandUpdateMatrix
			um(&mgc)
			gc = &mgc
		case CommandSendGarbage:
			var mgc GameCommandSendGarbage
			um(&mgc)
			gc = &mgc
		case CommandReceiveGarbage:
			var mgc GameCommandReceiveGarbage
			um(&mgc)
			gc = &mgc
		case CommandStats:
			var mgc GameCommandStats
			um(&mgc)
			gc = &mgc
		case CommandListGames:
			var mgc GameCommandListGames
			um(&mgc)
			gc = &mgc
		default:
			// TODO Require at least debug log level
			log.Println("unknown serverconn command", scanner.Text())
			continue
		}

		if !processed {
			s.addSourceID(gc)
			s.In <- gc
		}

		err = s.conn.SetReadDeadline(time.Now().Add(ConnTimeout))
		if err != nil {
			break
		}
	}

	s.Close()
}

func (s *Conn) handleWrite() {
	if s.conn == nil {
		for range s.out {
			s.Done()
		}
		return
	}

	var (
		msg GameCommandTransport
		j   []byte
		err error
	)

	for e := range s.out {
		if s.Terminated {
			s.Done()
			continue
		}

		msg = GameCommandTransport{Command: e.Command()}

		msg.Data, err = json.Marshal(e)
		if err != nil {
			log.Fatal(err)
		}

		j, err = json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}
		j = append(j, '\n')

		err = s.conn.SetWriteDeadline(time.Now().Add(ConnTimeout))
		if err != nil {
			s.Close()
		}

		_, err = s.conn.Write(j)
		if err != nil {
			s.Close()
		}

		s.LastTransfer = time.Now()
		s.conn.SetWriteDeadline(time.Time{})
		s.Done()
	}
}

func (s *Conn) Close() {
	if s.Terminated {
		return
	}

	s.Terminated = true

	s.conn.Close()

	go func() {
		s.Wait()
		close(s.In)
		close(s.out)
	}()
}

// When newGame is set to a ListedGame and gameID is 0, a new custom game is created
func (s *Conn) JoinGame(name string, gameID int, newGame *ListedGame, logger chan string, draw chan event.DrawObject) (*Game, error) {
	joinGameCommand := GameCommandJoinGame{Name: name, GameID: gameID}
	if newGame != nil {
		joinGameCommand.Listing.Name = newGame.Name
		joinGameCommand.Listing.MaxPlayers = newGame.MaxPlayers
		joinGameCommand.Listing.SpeedLimit = newGame.SpeedLimit
	}
	s.Write(&joinGameCommand)

	var (
		g   *Game
		err error
	)

	for e := range s.In {
		//log.Printf("Receive JoinGame command %+v", e)

		switch e.Command() {
		case CommandMessage:
			if p, ok := e.(*GameCommandMessage); ok {
				prefix := "* "
				if p.Player > 0 {
					name := "Anonymous"
					if player, ok := g.Players[p.Player]; ok {
						name = player.Name
					}
					prefix = "<" + name + "> "
				}

				if g != nil {
					g.Log(LogStandard, prefix+p.Message)
				} else {
					logger <- prefix + p.Message
				}
			}
		case CommandJoinGame:
			if p, ok := e.(*GameCommandJoinGame); ok {
				g, err = NewGame(4, s.Write, logger, draw)
				if err != nil {
					return nil, err
				}

				g.Lock()
				g.LocalPlayer = p.PlayerID
				g.Unlock()
			}
		case CommandUpdateGame:
			if g == nil {
				continue
			}

			if p, ok := e.(*GameCommandUpdateGame); ok {
				g.processUpdateGame(p)
			}
		case CommandStartGame:
			if p, ok := e.(*GameCommandStartGame); ok {
				if g != nil {
					g.Start(p.Seed)

					if p.Started {
						g.Players[g.LocalPlayer].Matrix.GameOver = true
					}

					go g.HandleReadCommands(s.In)

					return g, nil
				}
			}
		}
	}

	return nil, nil
}
