package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/game"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
	"git.sr.ht/~tslocum/netris/pkg/mino"

	"github.com/jroimartin/gocui"
	"github.com/mattn/go-isatty"
)

var (
	ready = make(chan bool)
	done  = make(chan bool)

	gm *game.Game
)

const RefreshRate = 15 * time.Millisecond

func init() {
	log.SetFlags(0)
}

func renderMatrix(m *matrix.Matrix) string {
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
	}

	return fmt.Sprintf("\033[38;5;%dm%c\033[0m", color, r)
}

func main() {
	flag.Parse()
	/*
		playerMatrix = matrix.NewMatrix(10, 20, 20)

		minos, err := mino.Generate(4)
		if err != nil {
			panic(err)
		}

		m := minos[3]

		m = m.Rotate(90)

		m = m.Rotate(90)

		m = m.Rotate(90)

		log.Println(m)
		log.Println()

		err = playerMatrix.Add(m, mino.BlockSolidCyan, mino.Point{rand.Intn(8), 4}, false)
		if err != nil {
			panic(err)
		}

		log.Println(m.Render())

		os.Exit(0)
	*/
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

	gm, err = game.NewGame(4, 123454)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			e := <-gm.Event

			fmt.Fprintln(dbg, e.Message)
		}
	}()

	gm.Start()

	go func() {
		for {
			time.Sleep(RefreshRate)

			gui.Update(func(i *gocui.Gui) error {
				gm.Lock()
				gm.Unlock()

				renderPreviewMatrix()
				renderPlayerMatrix()

				return nil
			})
		}
	}()

	// Game logic

	<-done

	if !closedGUI {
		closedGUI = true

		gui.Close()
	}
}
