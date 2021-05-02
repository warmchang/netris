package game

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.rocketnine.space/tslocum/netris/pkg/event"
	"code.rocketnine.space/tslocum/netris/pkg/mino"
)

const (
	UpdateDuration = 850 * time.Millisecond
	IdleStart      = 5 * time.Second
	IdleTimeout    = 1 * time.Minute
)

const (
	LogStandard = iota
	LogDebug
	LogVerbose
)

var Version = "0.0.0"

var gameNameRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-!@#$%^&*+=,./?()\[\]{};:<>'" ]+`)

type Game struct {
	ID   int
	Name string

	Starting           bool
	Started            bool
	TimeStarted        time.Time
	gameOver           bool
	sentGameOverMatrix bool

	Eternal    bool
	Terminated bool

	Local       bool
	LocalPlayer int
	nextPlayer  int
	Players     map[int]*Player
	MaxPlayers  int

	Event chan interface{}
	out   func(GameCommandInterface)

	draw     chan event.DrawObject
	logger   chan string
	LogLevel int

	Rank  int
	Minos []mino.Mino
	Seed  int64

	FallTime   time.Duration
	SpeedLimit int

	sentPing time.Time
	sync.Mutex
}

func NewGame(rank int, out func(GameCommandInterface), logger chan string, draw chan event.DrawObject) (*Game, error) {
	minos, err := mino.Generate(rank)
	if err != nil {
		return nil, err
	}

	g := &Game{
		Name:       "netris",
		Rank:       rank,
		Minos:      minos,
		nextPlayer: 1,
		Players:    make(map[int]*Player),
		Event:      make(chan interface{}, CommandQueueSize),
		draw:       draw,
		logger:     logger,
	}

	if out != nil {
		g.out = out
	} else {
		g.LocalPlayer = PlayerHost
		g.out = func(commandInterface GameCommandInterface) {
			// Do nothing
		}
	}

	g.FallTime = 850 * time.Millisecond

	go g.handleDropTerminatedPlayers()

	return g, nil
}

func (g *Game) Log(level int, a ...interface{}) {
	if g.logger == nil || level > g.LogLevel {
		return
	}

	g.logger <- fmt.Sprint(a...)
}

func (g *Game) Logf(level int, format string, a ...interface{}) {
	if g.logger == nil || level > g.LogLevel {
		return
	}

	g.logger <- fmt.Sprintf(format, a...)
}

func (g *Game) AddPlayer(p *Player) {
	g.Lock()
	defer g.Unlock()

	g.AddPlayerL(p)
}

func (g *Game) AddPlayerL(p *Player) {
	if p.Player == PlayerUnknown {
		if g.LocalPlayer != PlayerHost {
			return
		}

		p.Player = g.nextPlayer
		g.nextPlayer++
	}

	g.Players[p.Player] = p

	// TODO Verify rank-2 is valid for all playable rank previews
	p.Preview = mino.NewMatrix(g.Rank, g.Rank-2, 0, 1, g.Event, g.draw, mino.MatrixPreview)
	p.Preview.PlayerName = p.Name

	p.Matrix = mino.NewMatrix(10, 20, 4, 1, g.Event, g.draw, mino.MatrixStandard)
	p.Matrix.PlayerName = p.Name

	if g.Started {
		p.Matrix.SetGameOver()
	}

	if g.LocalPlayer == PlayerHost {
		p.Write(&GameCommandJoinGame{PlayerID: p.Player})

		var players = make(map[int]string)
		for _, player := range g.Players {
			players[player.Player] = player.Name
		}

		g.WriteAllL(&GameCommandUpdateGame{Players: players})

		if g.Started {
			p.Write(&GameCommandStartGame{Seed: g.Seed, Started: g.Started})
		}

		if len(g.Players) > 1 {
			g.WriteMessage(fmt.Sprintf("%s has joined the game", p.Name))
		}
	}
}

func (g *Game) RemovePlayer(playerID int) {
	g.Lock()
	defer g.Unlock()

	g.RemovePlayerL(playerID)
}

func (g *Game) RemovePlayerL(playerID int) {
	if playerID < 0 {
		return
	}

	p, ok := g.Players[playerID]
	if !ok || p == nil {
		return
	}

	playerName := p.Name

	delete(g.Players, playerID)

	if g.LocalPlayer == PlayerHost {
		if len(g.Players) == 0 {
			g.StopL()
			return
		}

		var players = make(map[int]string)
		for _, player := range g.Players {
			players[player.Player] = player.Name
		}

		g.WriteAllL(&GameCommandUpdateGame{Players: players})

		g.WriteMessage(fmt.Sprintf("%s has left the game", playerName))
	}
}

func (g *Game) WriteAll(gc GameCommandInterface) {
	g.Lock()
	defer g.Unlock()

	g.WriteAll(gc)
}

func (g *Game) WriteAllL(gc GameCommandInterface) {
	for i := range g.Players {
		g.Players[i].Write(gc)
	}
}

func (g *Game) WriteAllAndLogL(gc GameCommandInterface) {
	for i := range g.Players {
		g.Players[i].Write(gc)
	}
}

func (g *Game) WriteMessage(message string) {
	g.Log(LogStandard, message)
	g.WriteAllL(&GameCommandMessage{Player: PlayerHost, Message: message})
}

func (g *Game) Start(seed int64) int64 {
	g.Lock()
	defer g.Unlock()

	return g.StartL(seed)
}

func (g *Game) StartL(seed int64) int64 {
	restarting := g.Seed != 0

	if g.gameOver || g.Started {
		return g.Seed
	}

	g.Started = true
	g.TimeStarted = time.Now()

	if g.LocalPlayer == PlayerUnknown {
		log.Fatal("failed to start game: player unknown")
	}

	if seed == 0 {
		seed = time.Now().UTC().UnixNano()
	}
	g.Seed = seed

	for _, p := range g.Players {
		bag, err := mino.NewBag(g.Seed, g.Minos, 10)
		if err != nil {
			log.Fatalf("failed to start game: failed to create bag: %s", err)
		}

		p.Preview.AttachBag(bag)
		p.Matrix.AttachBag(bag)
	}

	// Take piece on host as well to give initial position for start of game
	for playerID, p := range g.Players {
		if !p.Matrix.TakePiece() {
			g.Log(LogStandard, "Failed to take piece while starting game for player ", p.Player)
			g.RemovePlayerL(playerID)
		}
	}

	if !restarting {
		go g.handle()
	}

	if g.LocalPlayer == PlayerHost {
		for i := range g.Players {
			g.Players[i].Write(&GameCommandStartGame{Seed: seed})
		}

		if !restarting {
			go g.handleDistributeMatrixes()
			go g.handleDistributeGarbage()
		}
	} else {
		if !restarting {
			go g.handleLowerPiece()

			go g.Players[g.LocalPlayer].Matrix.HandleReceiveGarbage()
			go g.handleSendMatrix()
		}
	}

	g.Logf(LogDebug, "Starting game %d", g.Seed)

	g.draw <- event.DrawAll

	return g.Seed
}

func (g *Game) Reset() {
	g.Lock()
	defer g.Unlock()

	g.ResetL()
}

func (g *Game) ResetL() {
	g.Log(LogDebug, "Resetting...")

	g.Starting = false
	g.Started = false
	g.TimeStarted = time.Time{}
	g.setGameOverL(false)
	g.sentGameOverMatrix = false

	for _, p := range g.Players {
		p.totalGarbageSent = 0
		p.totalGarbageReceived = 0
		p.pendingGarbage = 0
		p.Score = 0

		p.Preview.Reset()
		p.Matrix.Reset()
	}

	if g.LocalPlayer == PlayerHost {
		g.WriteAllL(&GameCommandJoinGame{})
	}

	g.draw <- event.DrawAll
}

func (g *Game) StopL() {
	if g.Terminated || g.Eternal {
		return
	}

	g.Terminated = true

	for playerID := range g.Players {
		g.RemovePlayerL(playerID)
	}
}

func (g *Game) handleSendMatrix() {
	m := g.Players[g.LocalPlayer].Matrix

	var matrixes = make(map[int]*mino.Matrix)

	t := time.NewTicker(UpdateDuration)
	for {
		<-t.C

		g.Lock()

		if !g.Started || (g.sentGameOverMatrix && m.GameOver) {
			g.Unlock()
			continue
		}

		matrixes[0] = m

		g.out(&GameCommandUpdateMatrix{Matrixes: matrixes})

		if m.GameOver {
			g.sentGameOverMatrix = true
		}

		g.Unlock()
	}
}

func (g *Game) handleDistributeMatrixes() {
	var matrixes map[int]*mino.Matrix
	t := time.NewTicker(UpdateDuration)
	for {
		<-t.C

		g.Lock()

		if g.Terminated {
			t.Stop()
			g.Unlock()
			return
		}

		remainingPlayer := -1
		remainingPlayers := 0

		for playerID, p := range g.Players {
			if !g.gameOver && !p.Matrix.GameOver && !g.Local && time.Since(p.Moved) >= IdleStart && time.Since(g.TimeStarted) >= IdleStart {
				p.Idle += UpdateDuration
				if p.Idle >= IdleTimeout {
					// Disconnect idle player
					p.Write(&GameCommandDisconnect{Player: playerID, Message: "Idling is not allowed"})
					g.RemovePlayerL(playerID)

					p := p
					go func(p *Player) {
						time.Sleep(time.Second)
						p.Close()
					}(p)
				}
			}

			if !g.gameOver && !p.Matrix.GameOver {
				remainingPlayer = playerID
				remainingPlayers++
			}
		}

		requiredPlayers := 2
		if g.Local {
			requiredPlayers = 1
		}

		if !g.gameOver && remainingPlayers < requiredPlayers {
			g.setGameOverL(true)

			if g.Local {
				g.WriteMessage("Game over")

				go func() {
					time.Sleep(3 * time.Second)

					g.Reset()
					g.Start(0)
				}()
			} else {
				winner := "Tie!"
				var (
					garbageSent []int
					players     []*Player
				)
				for i, p := range g.Players {
					p := p // Capture

					if i == remainingPlayer {
						winner = p.Name
					}

					garbageSent = append(garbageSent, p.totalGarbageSent)
					players = append(players, p)
				}
				sort.Slice(players, func(i, j int) bool {
					return garbageSent[i] < garbageSent[j]
				})

				g.WriteAllL(&GameCommandGameOver{Winner: winner})

				var garbageMessage strings.Builder
				for i, p := range players {
					if i > 0 {
						garbageMessage.WriteString(", ")
					}

					garbageMessage.WriteString(fmt.Sprintf("%s %d/%d", p.Name, p.totalGarbageSent, p.totalGarbageReceived))
				}

				g.WriteMessage(fmt.Sprintf("Winner: %s - Garbage sent/received: %s", winner, garbageMessage.String()))

				if len(g.Players) < 2 {
					g.WriteMessage("Game will start when there are at least two players")
				}

				go func() {
					for {
						time.Sleep(7 * time.Second)

						g.Lock()

						if g.Terminated {
							g.Unlock()
							return
						} else if len(g.Players) > 1 {
							g.Unlock()
							g.Reset()
							g.Start(0)
							return
						}

						g.Unlock()
					}
				}()
			}
		}

		matrixes = make(map[int]*mino.Matrix)
		for playerID, player := range g.Players {
			player.Matrix.PlayerName = player.Name
			player.Matrix.GarbageSent = player.totalGarbageSent
			player.Matrix.GarbageReceived = player.totalGarbageReceived

			matrixes[playerID] = player.Matrix
		}
		g.WriteAllL(&GameCommandUpdateMatrix{Matrixes: matrixes})

		g.Unlock()
	}
}

func (g *Game) HandleReadCommands(in chan GameCommandInterface) {
	var e GameCommandInterface
	for e = range in {
		g.Lock()

		c := e.Command()

		logLevel := LogDebug
		if c == CommandPing || c == CommandPong || c == CommandUpdateMatrix {
			logLevel = LogVerbose
		}
		g.Log(logLevel, "LOCAL handle ", e.Command(), " from ", e.Source(), " ", e)

		switch e.Command() {
		case CommandDisconnect:
			if p, ok := e.(*GameCommandDisconnect); ok {
				if p.Player == g.LocalPlayer {
					if p.Message != "" {
						g.Logf(LogStandard, "* Disconnected - Reason: %s", p.Message)
					} else {
						g.Logf(LogStandard, "* Disconnected")
					}

					g.setGameOverL(true)
				}
			}
		case CommandPong:
			if p, ok := e.(*GameCommandPong); ok {
				if len(p.Message) > 1 && p.Message[0] == 'm' {
					if i, err := strconv.ParseInt(p.Message[1:], 10, 64); err == nil {
						if i == g.sentPing.UnixNano() {
							g.Logf(LogStandard, "* Server latency is %dms", time.Since(g.sentPing).Milliseconds())

							g.sentPing = time.Time{}
						}
					}
				}
			}
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
				g.Log(LogStandard, prefix+p.Message)
			}
		case CommandNickname:
			if p, ok := e.(*GameCommandNickname); ok {
				if player, ok := g.Players[p.Player]; ok {
					newNick := Nickname(p.Nickname)
					if newNick != "" && newNick != player.Name {
						oldNick := player.Name

						player.Name = newNick

						if p.Player == g.LocalPlayer {
							g.Players[g.LocalPlayer].Matrix.PlayerName = newNick
						}

						g.Logf(LogStandard, "* %s is now known as %s", oldNick, newNick)
					}
				}
			}
		case CommandJoinGame:
			g.ResetL()
		case CommandQuitGame:
			if p, ok := e.(*GameCommandQuitGame); ok {
				g.RemovePlayerL(p.Player)
			}
		case CommandUpdateGame:
			if p, ok := e.(*GameCommandUpdateGame); ok {
				g.processUpdateGameL(p)
			}
		case CommandStartGame:
			if p, ok := e.(*GameCommandStartGame); ok {
				g.StartL(p.Seed)

				if p.Started {
					g.Players[g.LocalPlayer].Matrix.SetGameOver()
				}
			}
		case CommandUpdateMatrix:
			if p, ok := e.(*GameCommandUpdateMatrix); ok {
				for player, m := range p.Matrixes {
					if player == g.LocalPlayer {
						g.Players[player].Matrix.GarbageSent = m.GarbageSent
						g.Players[player].Matrix.GarbageReceived = m.GarbageReceived

						continue
					} else if _, ok := g.Players[player]; !ok {
						continue
					}

					g.Players[player].Matrix.Replace(m)
				}

				g.draw <- event.DrawMultiplayerMatrixes
			}
		case CommandReceiveGarbage:
			if p, ok := e.(*GameCommandReceiveGarbage); ok {
				g.Players[g.LocalPlayer].Matrix.AddPendingGarbage(p.Lines)
			}
		case CommandGameOver:
			if p, ok := e.(*GameCommandGameOver); ok {
				if p.Winner != "" {
					g.setGameOverL(true)
				} else {
					g.Players[p.Player].Matrix.SetGameOver()

					g.draw <- event.DrawMultiplayerMatrixes
				}
			}
		case CommandStats:
			if p, ok := e.(*GameCommandStats); ok {
				g.Logf(LogStandard, "* %d players in %d games - uptime: %s", p.Players, p.Games, time.Since(p.Created.Local()).Truncate(time.Minute))

			}
		default:
			g.Log(LogStandard, "unknown handle read command", e.Command(), e)
		}

		g.Unlock()
	}
}

func (g *Game) setGameOverL(gameOver bool) {
	if g.gameOver == gameOver {
		return
	}

	g.gameOver = gameOver

	if g.gameOver {
		for _, p := range g.Players {
			p.Matrix.SetGameOver()
		}

		g.draw <- event.DrawAll
	}
}

func (g *Game) handleDistributeGarbage() {
	t := time.NewTicker(500 * time.Millisecond)
	for {
		<-t.C

		g.Lock()

		if g.Terminated {
			t.Stop()
			g.Unlock()
			return
		}

		for i := range g.Players {
			if g.Players[i].pendingGarbage > 0 {
				g.Players[i].Write(&GameCommandReceiveGarbage{Lines: g.Players[i].pendingGarbage})

				g.Players[i].pendingGarbage = 0
			}
		}

		g.Unlock()
	}
}

func (g *Game) handle() {
	var e interface{}
	for {
		e = <-g.Event

		g.Log(LogDebug, "Game handle", e)

		if ev, ok := e.(*event.MessageEvent); ok {
			g.out(&GameCommandMessage{Message: ev.Message})
		} else if _, ok := e.(*event.GameOverEvent); ok {
			g.Players[g.LocalPlayer].Matrix.SetGameOver()

			g.out(&GameCommandGameOver{})
		} else if ev, ok := e.(*event.NicknameEvent); ok {
			g.out(&GameCommandNickname{Nickname: ev.Nickname})
		} else if ev, ok := e.(*event.SendGarbageEvent); ok {
			g.out(&GameCommandSendGarbage{Lines: ev.Lines})
		} else if ev, ok := e.(*event.ScoreEvent); ok {
			g.Players[ev.Player].Score += ev.Score

			if ev.Message != "" {
				g.Log(LogStandard, ev.Message)
			}
		} else if ev, ok := e.(*event.Event); ok {
			if ev.Message != "" {
				g.Log(LogStandard, ev.Message)
			}
		} else {
			log.Fatalf("unknown event type: %v", e)
		}
	}
}

func (g *Game) handleLowerPiece() {
	var (
		ticker *time.Ticker
	)

	m := g.Players[g.LocalPlayer].Matrix

	ticker = time.NewTicker(g.FallTime)
	for {
		select {
		case <-m.Move:
			ticker.Stop()
			ticker = time.NewTicker(g.FallTime)
			continue
		case <-ticker.C:
			for {
				select {
				case <-m.Move:
					continue
				default:
				}

				break
			}
		}

		g.Lock()
		m.LowerPiece()
		g.Unlock()
	}
}

func (g *Game) processUpdateGame(gc *GameCommandUpdateGame) {
	g.Lock()
	defer g.Unlock()

	g.processUpdateGameL(gc)
}

func (g *Game) processUpdateGameL(gc *GameCommandUpdateGame) {
	for playerID, playerName := range gc.Players {
		if existingPlayer, ok := g.Players[playerID]; ok {
			existingPlayer.Name = playerName
		} else {
			pl := NewPlayer(playerName, nil)
			pl.Player = playerID

			g.AddPlayerL(pl)
		}
	}
	for playerID := range g.Players {
		if _, ok := gc.Players[playerID]; !ok {
			g.RemovePlayerL(playerID)
		}
	}

	g.draw <- event.DrawMultiplayerMatrixes
}

func (g *Game) ProcessAction(a event.GameAction) {
	g.Lock()
	defer g.Unlock()

	g.ProcessActionL(a)
}

func (g *Game) ProcessActionL(a event.GameAction) {
	if p, ok := g.Players[g.LocalPlayer]; ok {
		if p.Matrix == nil {
			return
		}

		switch a {
		case event.ActionRotateCCW:
			p.Matrix.RotatePiece(1, 1)
		case event.ActionRotateCW:
			p.Matrix.RotatePiece(1, 0)
		case event.ActionMoveLeft:
			p.Matrix.MovePiece(-1, 0)
		case event.ActionMoveRight:
			p.Matrix.MovePiece(1, 0)
		case event.ActionSoftDrop:
			p.Matrix.MovePiece(0, -1)
		case event.ActionHardDrop:
			p.Matrix.HardDropPiece()
		case event.ActionNick:
			g.out(&GameCommandNickname{Nickname: Nickname(p.Name)})
		case event.ActionPing:
			g.sentPing = time.Now()
			g.out(&GameCommandPing{Message: fmt.Sprintf("m%d", g.sentPing.UnixNano())})
		case event.ActionStats:
			g.out(&GameCommandStats{})
		}
	}
}

func (g *Game) handleDropTerminatedPlayers() {
	for {
		time.Sleep(15 * time.Second)

		g.Lock()

		if g.Terminated {
			g.Unlock()
			return
		}

		for playerID, p := range g.Players {
			if p.Terminated {
				g.RemovePlayerL(playerID)
			}
		}

		g.Unlock()
	}
}

func GameName(name string) string {
	name = gameNameRegexp.ReplaceAllString(strings.TrimSpace(name), "")
	if len(name) > 24 {
		name = name[:24]
	} else if name == "" {
		name = "netris"
	}

	return name
}
