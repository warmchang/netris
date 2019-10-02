package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/player"

	"git.sr.ht/~tslocum/netris/pkg/game"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
	"github.com/mattn/go-isatty"
)

var (
	ready = make(chan bool)
	done  = make(chan bool)

	activeGame *game.Game

	connectAddress string
	debugAddress   string
)

const RefreshRate = 20 * time.Millisecond

func init() {
	log.SetFlags(0)
}

func renderMatrix(m *mino.Matrix) string {
	var b strings.Builder

	for y := m.H - 1; y >= 0; y-- {
		for x := 0; x < m.W; x++ {
			b.WriteString(renderBlock(m.Block(x, y)))
		}

		if y == 0 {
			break
		}

		b.WriteRune('\n')
	}

	return b.String()
}

func renderBlock(b mino.Block) string {
	r := b.Rune()

	color := 39

	switch b {
	case mino.BlockGhostBlue, mino.BlockSolidBlue:
		color = 25
	case mino.BlockGhostCyan, mino.BlockSolidCyan:
		color = 45
	case mino.BlockGhostRed, mino.BlockSolidRed:
		color = 160
	case mino.BlockGhostYellow, mino.BlockSolidYellow:
		color = 226
	case mino.BlockGhostMagenta, mino.BlockSolidMagenta:
		color = 91
	case mino.BlockGhostGreen, mino.BlockSolidGreen:
		color = 46
	case mino.BlockGhostOrange, mino.BlockSolidOrange:
		color = 202
	case mino.BlockGarbage:
		color = 8
	}

	return fmt.Sprintf("\033[38;5;%dm%c\033[0m", color, r)
}

func main() {
	flag.StringVar(&connectAddress, "connect", "", "server address to connect to")
	flag.StringVar(&debugAddress, "debug", "", "address to serve debug info")
	flag.Parse()

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	var err error

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start netris: non-interactive terminals are not supported")
	}

	err = initGUI()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Fatal(err)
		}

		done <- true
	}()

	<-ready

	if connectAddress != "" {
		s := game.ConnectUnix("/tmp/netris.sock")

		s.Out <- player.GameCommand{C: player.CommandJoinGame}

		select {}

		panic(fmt.Sprint("connected", s))
	}

	go func() {
		for {
			time.Sleep(RefreshRate)

			gui.Update(func(i *gocui.Gui) error {
				renderPreviewMatrix()
				renderPlayerMatrix()

				return nil
			})
		}
	}()

	activeGame, err = game.NewGame(4, 4)

	server, client := net.Pipe()
	_ = server

	p := player.NewPlayer(client)

	activeGame.AddPlayer(p)

	activeGame.Start()

	// Game logic

	<-done

	if !closedGUI {
		closedGUI = true

		gui.Close()
	}
}
