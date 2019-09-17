package main

import (
	"fmt"
	"sync"

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
)

func initGUI() error {
	var err error
	gui, err = gocui.NewGui(gocui.OutputNormal)
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
	if v, err := gui.SetView("info", 14, 3, 20, 8); err != nil {
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
