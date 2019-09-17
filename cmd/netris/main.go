package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.sr.ht/~tslocum/netris/pkg/mino"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
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

func main() {
	flag.Parse()

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start gmenu: non-interactive terminals are not supported")
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

	n := matrix.NewMatrix(4, 4, 0)

	n.M[n.I(0, 0)] = mino.BlockSolid
	n.M[n.I(1, 0)] = mino.BlockSolid
	n.M[n.I(2, 0)] = mino.BlockSolid
	n.M[n.I(3, 0)] = mino.BlockSolid

	fmt.Fprint(info, n.Render())

	m := matrix.NewMatrix(10, 20, 20)
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
	fmt.Fprint(mtx, m.Render())

	<-done

	if !closedGUI {
		closedGUI = true

		gui.Close()
	}
}
