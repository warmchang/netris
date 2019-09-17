package main

import (
	"fmt"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/jroimartin/gocui"
)

func moveLeft(_ *gocui.Gui, _ *gocui.View) error {
	fmt.Fprintln(info, "left")
	return nil
}

func moveRight(_ *gocui.Gui, _ *gocui.View) error {
	fmt.Fprintln(info, "right")
	return nil
}

func moveUp(_ *gocui.Gui, _ *gocui.View) error {
	minos, err := mino.Generate(4)
	if err != nil {
		panic(err)
	}

	if playerBag == nil {
		playerBag = mino.NewBag(minos)
	}

	setNextPiece(playerBag.Take())

	return nil
}

func moveDown(_ *gocui.Gui, _ *gocui.View) error {
	fmt.Fprintln(info, "down")
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
