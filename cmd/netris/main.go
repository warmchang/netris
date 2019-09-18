package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
	"git.sr.ht/~tslocum/netris/pkg/mino"

	"github.com/jroimartin/gocui"
	"github.com/mattn/go-isatty"
)

var (
	ready = make(chan bool)
	done  = make(chan bool)
)

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

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start netris: non-interactive terminals are not supported")
	}

	err := initGUI()
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

	/*m := matrix.NewMatrix(10, 20, 20)
	m.M[m.I(5, 20)] = mino.BlockSolid
	m.M[m.I(4, 21)] = mino.BlockSolid
	m.M[m.I(5, 21)] = mino.BlockSolid
	m.M[m.I(6, 21)] = mino.BlockSolid

	m.M[m.I(5, 38)] = mino.BlockGhost
	m.M[m.I(4, 39)] = mino.BlockGhost
	m.M[m.I(5, 39)] = mino.BlockGhost
	m.M[m.I(6, 39)] = mino.BlockGhost

	m.M[m.I(8, 38)] = mino.BlockSolid
	m.M[m.I(9, 38)] = mino.BlockSolid
	m.M[m.I(8, 39)] = mino.BlockSolid
	m.M[m.I(9, 39)] = mino.BlockSolid

	mtx.Clear()
	fmt.Fprint(mtx, m.Render())*/

	<-done

	if !closedGUI {
		closedGUI = true

		gui.Close()
	}
}
