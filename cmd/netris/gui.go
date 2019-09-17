package main

import (
	"fmt"
	"math/rand"
	"sync"

	"git.sr.ht/~tslocum/netris/pkg/matrix"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
)

var (
	gui       *gocui.Gui
	closedGUI bool

	info   *gocui.View
	mtx    *gocui.View
	buffer *gocui.View

	bufferActive bool

	initialDraw sync.Once

	playerMatrix   *matrix.Matrix
	playerBag      *mino.Bag
	newPieceMatrix *matrix.Matrix
)

func initGUI() error {
	var err error
	gui, err = gocui.NewGui(gocui.Output256)
	if err != nil {
		return err
	}

	gui.InputEsc = true
	gui.Cursor = true
	gui.Mouse = false

	gui.SetManagerFunc(layout)

	if err := keybindings(); err != nil {
		return err
	}

	return nil
}

func layout(_ *gocui.Gui) error {
	maxX, maxY := gui.Size()
	listWidth := maxX

	if v, err := gui.SetView("matrix", 1, 2, 12, 23); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		mtx = v

		v.Frame = true
		v.Wrap = false
	}
	if v, err := gui.SetView("info", 14, 3, 42, 10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		info = v

		v.Frame = false
		v.Wrap = false
	}
	if v, err := gui.SetView("buffer", -1, -1, listWidth, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		buffer = v

		v.Frame = false
		v.Wrap = false
		v.Editable = bufferActive
		v.Wrap = true
		//v.Editor = gocui.EditorFunc(searchEditor)

		if _, err := gui.SetCurrentView("buffer"); err != nil {
			return err
		}
	}

	initialDraw.Do(func() {
		bufferActive = true // Force draw
		setBufferStatus(false)

		ready <- true
	})

	_, _ = maxX, maxY
	return nil
}

func setBufferStatus(active bool) {
	if bufferActive == active {
		return
	}

	bufferActive = active

	buffer.Editable = active

	buffer.Clear()

	if active {
		buffer.SetCursor(0, 0)
		gui.Cursor = true

		return
	}

	gui.Cursor = false
	printHeader()
}

func printHeader() {
	if bufferActive {
		return
	}

	fmt.Fprintln(buffer, "Welcome to netris")
}

func setNextPiece(m mino.Mino) {
	pieceColor := mino.BlockSolidYellow
	switch m.String() {
	case mino.TetrominoI:
		pieceColor = mino.BlockSolidCyan
	case mino.TetrominoJ:
		pieceColor = mino.BlockSolidBlue
	case mino.TetrominoL:
		pieceColor = mino.BlockSolidOrange
	case mino.TetrominoO:
		pieceColor = mino.BlockSolidYellow
	case mino.TetrominoS:
		pieceColor = mino.BlockSolidGreen
	case mino.TetrominoT:
		pieceColor = mino.BlockSolidMagenta
	case mino.TetrominoZ:
		pieceColor = mino.BlockSolidRed
	}

	rank := len(m)
	if newPieceMatrix == nil || newPieceMatrix.W < rank || newPieceMatrix.H < rank {
		newPieceMatrix = matrix.NewMatrix(rank, rank, 0)
	}

	newPieceMatrix.Clear()
	err := newPieceMatrix.Add(m, pieceColor, mino.Point{0, 0})
	if err != nil {
		panic(err)
	}

	info.Clear()
	fmt.Fprint(info, renderMatrix(newPieceMatrix))

	//fmt.Fprint(info, "\n"+m.String())

	if playerMatrix == nil {
		playerMatrix = matrix.NewMatrix(10, 20, 20)
	}

RANDOMPIECE:
	for i := 0; i < playerMatrix.H; i++ {
		for j := 0; j < 300; j++ {
			err = playerMatrix.Add(m, pieceColor, mino.Point{rand.Intn(8), i})
			if err == nil {
				break RANDOMPIECE
			}
		}
	}

	renderPlayerMatrix()
}

func renderPlayerMatrix() {
	mtx.Clear()

	if playerMatrix == nil {
		return
	}

	fmt.Fprint(mtx, renderMatrix(playerMatrix))
}

func closeGUI() {
	if closedGUI {
		return
	}
	closedGUI = true

	gui.Close()

	gui.Update(func(_ *gocui.Gui) error {
		return gocui.ErrQuit
	})
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	closeGUI()
	return gocui.ErrQuit
}
