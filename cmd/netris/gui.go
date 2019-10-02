package main

import (
	"fmt"
	"strconv"
	"sync"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
)

var (
	gui       *gocui.Gui
	closedGUI bool

	info   *gocui.View
	mtx    *gocui.View
	input  *gocui.View
	buffer *gocui.View

	inputActive bool

	initialDraw sync.Once
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

	if v, err := gui.SetView("matrix", 1, 1, 12, 22); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		mtx = v

		v.Frame = true
		v.Wrap = false
	}
	if v, err := gui.SetView("info", 14, 3, 24, 20); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		info = v

		v.Frame = false
		v.Wrap = false
	}

	// TODO: Remove.  When enter is pressed, the chat history is shown instead of other players matrixes
	// When there are no other matrixes, history is always displayed
	if v, err := gui.SetView("buffer", 24, 1, maxX, maxY+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		buffer = v

		v.Frame = false
		v.Wrap = true
		v.Autoscroll = true
	}

	if v, err := gui.SetView("input", -1, -1, listWidth, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		input = v

		v.Frame = false
		v.Wrap = false
		v.Editable = inputActive
		v.Wrap = true
		//v.Editor = gocui.EditorFunc(searchEditor)

		if _, err := gui.SetCurrentView("input"); err != nil {
			return err
		}
	}

	initialDraw.Do(func() {
		inputActive = true // Force draw
		setInputStatus(false)

		ready <- true
	})

	_, _ = maxX, maxY
	return nil
}

func setInputStatus(active bool) {
	if inputActive == active {
		return
	}

	inputActive = active

	input.Editable = active

	input.Clear()

	if active {
		input.SetCursor(0, 0)
		gui.Cursor = true

		return
	}

	gui.Cursor = false
	printHeader()
}

func printHeader() {
	if inputActive {
		return
	}

	fmt.Fprintln(input, "Welcome to netris")
}

func renderPreviewMatrix() {
	info.Clear()

	g := activeGame
	if g == nil {
		return
	}

	m := mino.NewPiece(g.Bags[0].Next(), mino.Point{0, 0})

	solidBlock := m.SolidBlock()

	g.Previews[0].Clear()

	err := g.Previews[0].Add(m, solidBlock, mino.Point{0, 0}, false)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(info, renderMatrix(g.Previews[0]))
	fmt.Fprint(info, "\n\n\n\n\n\nScore:\n\n"+strconv.Itoa(g.Scores[0]))
}

func renderPlayerMatrix() {
	mtx.Clear()

	g := activeGame
	if g == nil {
		return
	}

	p := g.Matrixes[0].P[0]

	ghostBlock := p.GhostBlock()
	solidBlock := p.SolidBlock()

	g.Matrixes[0].ClearOverlay()
	if p != nil {
		// Draw ghost piece
		for y := p.Y; y >= 0; y-- {
			if y == 0 || !g.Matrixes[0].CanAddAt(p, mino.Point{p.X, y - 1}) {
				err := g.Matrixes[0].Add(p, ghostBlock, mino.Point{p.X, y}, true)
				if err != nil {
					panic(fmt.Sprintf("failed to draw ghost piece: %+v", err))
				}

				break
			}
		}

		// Draw piece
		err := g.Matrixes[0].Add(p, solidBlock, mino.Point{p.X, p.Y}, true)
		if err != nil {
			panic(fmt.Sprintf("failed to draw active piece: %+v", err))
		}

	}

	if g.Matrixes[0] == nil {
		return
	}

	fmt.Fprint(mtx, renderMatrix(g.Matrixes[0]))
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
