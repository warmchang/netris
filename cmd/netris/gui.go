package main

import (
	"fmt"
	"sync"

	"git.sr.ht/~tslocum/netris/pkg/game"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
)

var (
	gui       *gocui.Gui
	closedGUI bool

	info   *gocui.View
	mtx    *gocui.View
	buffer *gocui.View
	dbg    *gocui.View

	bufferActive bool

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

	if v, err := gui.SetView("debug", 1, 24, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		dbg = v

		v.Frame = false
		v.Wrap = true
		v.Autoscroll = true
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

func printDebug(msg string) {
	gm.Event <- &game.Event{msg}
}

func printDebugf(format string, a ...interface{}) {
	printDebug(fmt.Sprintf(format+"\n", a...))
}

func printHeader() {
	if bufferActive {
		return
	}

	fmt.Fprintln(buffer, "Welcome to netris")
}

func renderPreviewMatrix() {
	m := gm.NextPieces[0]

	solidBlock := m.SolidBlock()

	gm.Previews[0].ClearMatrix()
	err := gm.Previews[0].Add(m, solidBlock, mino.Point{0, 0}, false)
	if err != nil {
		panic(err)
	}

	info.Clear()
	fmt.Fprint(info, renderMatrix(gm.Previews[0]))
}

func renderPlayerMatrix() {
	gm.Lock()
	defer gm.Unlock()

	mtx.Clear()

	ghostBlock := gm.Pieces[0].GhostBlock()
	solidBlock := gm.Pieces[0].SolidBlock()

	gm.Matrixes[0].ClearOverlay()
	if gm.Pieces[0] != nil {
		// Draw ghost piece
		for y := gm.Pieces[0].Y; y >= 0; y-- {
			if y == 0 || !gm.Matrixes[0].CanAddAt(gm.Pieces[0], mino.Point{gm.Pieces[0].X, y - 1}) {
				err := gm.Matrixes[0].Add(gm.Pieces[0], ghostBlock, mino.Point{gm.Pieces[0].X, y}, true)
				if err != nil {
					panic(err)
				}

				break
			}
		}

		// Draw piece
		err := gm.Matrixes[0].Add(gm.Pieces[0], solidBlock, mino.Point{gm.Pieces[0].X, gm.Pieces[0].Y}, true)
		if err != nil {
			panic(err)
		}

	}

	if gm.Matrixes[0] == nil {
		return
	}

	fmt.Fprint(mtx, renderMatrix(gm.Matrixes[0]))
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
