package main

import (
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
)

func moveLeft(_ *gocui.Gui, _ *gocui.View) error {
	gm.Lock()

	x := gm.Pieces[0].X - 1
	if x < 0 {
		x = 0
	}

	gm.Pieces[0].X = x
	gm.Unlock()
	renderPlayerMatrix()

	return nil
}

func moveRight(_ *gocui.Gui, _ *gocui.View) error {
	gm.Lock()

	x := gm.Pieces[0].X + 1
	if x+gm.Pieces[0].Width() >= gm.Matrixes[0].W {
		x = gm.Matrixes[0].W - gm.Pieces[0].Width()
	}

	gm.Pieces[0].X = x
	gm.Unlock()
	renderPlayerMatrix()

	return nil
}

func moveUp(_ *gocui.Gui, _ *gocui.View) error {
	gm.LandPiece(0)

	renderPlayerMatrix()

	return nil
}

func moveDown(_ *gocui.Gui, _ *gocui.View) error {
	gm.Lock()

	y := gm.Pieces[0].Y - 1
	if y >= 0 && gm.Matrixes[0].CanAdd(gm.Pieces[0], mino.Point{gm.Pieces[0].X, y}) {
		gm.Pieces[0].Y = y
	}

	gm.Unlock()
	gm.DroppedPiece(0)

	renderPlayerMatrix()

	return nil
}

func rotateBack(_ *gocui.Gui, _ *gocui.View) error {
	gm.Lock()
	gm.Pieces[0].Rotate(270)
	gm.Unlock()

	renderPlayerMatrix()

	return nil
}

func rotateForward(_ *gocui.Gui, _ *gocui.View) error {
	gm.Lock()
	gm.Pieces[0].Rotate(90)
	gm.Unlock()

	renderPlayerMatrix()

	return nil
}

func pressSelect(_ *gocui.Gui, _ *gocui.View) error {
	if bufferActive {
		// Process input
	}

	setBufferStatus(!bufferActive)

	return nil
}

func pressBack(_ *gocui.Gui, _ *gocui.View) error {
	if !bufferActive {
		return nil
	}

	setBufferStatus(false)

	return nil
}

func keybindings() error {
	if err := gui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, moveLeft); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, moveRight); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, moveUp); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, moveDown); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", 'Z', gocui.ModNone, rotateBack); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", 'z', gocui.ModNone, rotateBack); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", 'X', gocui.ModNone, rotateForward); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", 'x', gocui.ModNone, rotateForward); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, pressSelect); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, pressBack); err != nil {
		return err
	}

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}
