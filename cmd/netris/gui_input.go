package main

import (
	"github.com/jroimartin/gocui"
)

func moveLeft(_ *gocui.Gui, _ *gocui.View) error {
	gm.MovePiece(0, -1, 0)

	renderPlayerMatrix()

	return nil
}

func moveRight(_ *gocui.Gui, _ *gocui.View) error {
	gm.MovePiece(0, 1, 0)

	renderPlayerMatrix()

	return nil
}

func moveUp(_ *gocui.Gui, _ *gocui.View) error {
	gm.LandPiece(0)

	renderPlayerMatrix()

	return nil
}

func moveDown(_ *gocui.Gui, _ *gocui.View) error {
	gm.MovePiece(0, 0, -1)

	gm.DroppedPiece(0)

	renderPlayerMatrix()

	return nil
}

func rotateBack(_ *gocui.Gui, _ *gocui.View) error {
	gm.RotatePiece(0, 270)

	renderPlayerMatrix()

	return nil
}

func rotateForward(_ *gocui.Gui, _ *gocui.View) error {
	printDebug("F1 " + gm.Pieces[0].String())
	gm.RotatePiece(0, 90)

	printDebug("F2 " + gm.Pieces[0].String())
	renderPlayerMatrix()
	printDebug("F3 " + gm.Pieces[0].String())

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
