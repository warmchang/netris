package main

import (
	"github.com/jroimartin/gocui"
)

func moveLeft(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].MovePiece(0, -1, 0)

	return nil
}

func moveRight(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].MovePiece(0, 1, 0)

	return nil
}

func moveUp(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].HardDropPiece(0)

	return nil
}

func moveDown(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].MovePiece(0, 0, -1)

	return nil
}

func rotateBack(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].Rotate(0, 1, 1)

	return nil
}

func rotateForward(_ *gocui.Gui, _ *gocui.View) error {
	if activeGame == nil {
		return nil
	}

	activeGame.Matrixes[0].Rotate(0, 1, 0)

	return nil
}

func pressSelect(_ *gocui.Gui, _ *gocui.View) error {
	if inputActive {
		// Process input
	}

	setInputStatus(!inputActive)

	return nil
}

func pressBack(_ *gocui.Gui, _ *gocui.View) error {
	if !inputActive {
		return nil
	}

	setInputStatus(false)

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
