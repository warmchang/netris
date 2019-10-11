package game

import (
	"bufio"
	"encoding/json"
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

type ServerConn struct {
	Conn   net.Conn
	Player int

	In         chan GameCommandInterface
	out        chan GameCommandInterface
	ForwardOut chan GameCommandInterface

	Terminated bool

	*sync.WaitGroup
}

func NewServerConn(conn net.Conn, forwardOut chan GameCommandInterface) *ServerConn {
	s := ServerConn{Conn: conn, WaitGroup: new(sync.WaitGroup)}

	s.In = make(chan GameCommandInterface, CommandQueueSize)
	s.out = make(chan GameCommandInterface, CommandQueueSize)
	s.ForwardOut = forwardOut

	if conn == nil {
		// Local instance

		go s.handleLocalWrite()
	} else {
		go s.handleRead()
		go s.handleWrite()
	}

	return &s
}

func ConnectUnix(path string) *ServerConn {
	tries := 0
	for {
		conn, err := net.DialTimeout("unix", path, ConnTimeout)
		if err != nil {
			if tries > 25 {
				log.Fatal("Listen error: ", err)
			} else {
				time.Sleep(250 * time.Millisecond)

				tries++
				continue
			}
		}

		return NewServerConn(conn, nil)
	}
}

func (s *ServerConn) Write(gc GameCommandInterface) {
	if s == nil || s.Terminated {
		return
	}

	s.Add(1)
	s.out <- gc
}

func (s *ServerConn) handleLocalWrite() {
	for e := range s.out {
		if s.ForwardOut != nil {
			s.ForwardOut <- e
		}

		s.Done()
	}
}

func (s *ServerConn) addSourceID(gc GameCommandInterface) {
	gc.SetSource(s.Player)
}

func (s *ServerConn) handleRead() {
	if s.Conn == nil {
		return
	}

	err := s.Conn.SetReadDeadline(time.Now().Add(ConnTimeout))
	if err != nil {
		s.Close()
		return
	}

	var (
		msg GameCommandTransport
		gc  GameCommandInterface
	)
	scanner := bufio.NewScanner(s.Conn)
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			panic(err)
		}

		if msg.Command == CommandMessage {
			var gameCommand GameCommandMessage
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandJoinGame {
			var gameCommand GameCommandJoinGame
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandQuitGame {
			var gameCommand GameCommandQuitGame
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandUpdateGame {
			var gameCommand GameCommandUpdateGame
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandStartGame {
			var gameCommand GameCommandStartGame
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandGameOver {
			var gameCommand GameCommandGameOver
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandUpdateMatrix {
			var gameCommand GameCommandUpdateMatrix
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandSendGarbage {
			var gameCommand GameCommandSendGarbage
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else if msg.Command == CommandReceiveGarbage {
			var gameCommand GameCommandReceiveGarbage
			err := json.Unmarshal(msg.Data, &gameCommand)
			if err != nil {
				panic(err)
			}

			gc = &gameCommand
		} else {
			log.Println("unknown serverconn command", scanner.Text())
			continue
		}

		s.addSourceID(gc)
		s.In <- gc

		err = s.Conn.SetReadDeadline(time.Now().Add(ConnTimeout))
		if err != nil {
			s.Close()
			return
		}
	}
}

func (s *ServerConn) handleWrite() {
	if s.Conn == nil {
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
		if p, ok := e.(*GameCommandMessage); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandJoinGame); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandQuitGame); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandUpdateGame); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandStartGame); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandGameOver); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandUpdateMatrix); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandSendGarbage); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else if p, ok := e.(*GameCommandReceiveGarbage); ok {
			msg.Data, err = json.Marshal(p)
			if err != nil {
				panic(err)
			}
		} else {
			log.Println(e.Command(), e)
			panic("unknown serverconn write command")
		}

		j, err = json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		j = append(j, '\n')

		err = s.Conn.SetWriteDeadline(time.Now().Add(ConnTimeout))
		if err != nil {
			s.Close()
		}

		_, err = s.Conn.Write(j)
		if err != nil {
			s.Close()
		}

		err = s.Conn.SetWriteDeadline(time.Time{})

		s.Done()
	}
}

func (s *ServerConn) Close() {
	if s.Terminated {
		return
	}

	s.Terminated = true

	go func() {
		s.Conn.Close()
		s.Wait()
		close(s.In)
		close(s.out)
	}()
}

func (s *ServerConn) JoinGame(name string, gameID int, logger chan string, draw chan event.DrawObject) (*Game, error) {
	s.Write(&GameCommandJoinGame{Name: name, GameID: gameID})
	var (
		g   *Game
		err error
	)

	for e := range s.In {
		//log.Printf("Receive JoinGame command %+v", e)

		switch e.Command() {
		case CommandMessage:
			if p, ok := e.(*GameCommandMessage); ok {
				if g != nil {
					prefix := "* "
					if p.Player > 0 {
						name := "Anonymous"
						if player, ok := g.Players[p.Player]; ok {
							name = player.Name
						}
						prefix = "<" + name + "> "
					}
					g.Log(LogStandard, prefix+p.Message)
				} else {
					logger <- p.Message
					draw <- event.DrawMessages
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
				// TODO Unify with JoinGame player update
				g.Lock()
				for playerID, playerName := range p.Players {
					if existingPlayer, ok := g.Players[playerID]; ok {
						existingPlayer.Name = playerName
					} else {
						pl := NewPlayer(playerName, nil)
						pl.Player = playerID

						g.AddPlayerL(pl)
					}
				}
				g.Unlock()
			} else {
				log.Println(e.Command(), " - ", e)
				panic("unknown payload")
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
		case CommandUpdateMatrix:
			// Do nothing (missed join game command)
		default:
			log.Println("unnknown joingame command", e.Command(), e)
		}
	}

	return nil, nil
}
