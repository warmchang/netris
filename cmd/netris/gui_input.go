package main

import "github.com/gdamore/tcell"

func handleKeypress(event *tcell.EventKey) *tcell.EventKey {
	if inputActive {
		if event.Key() == tcell.KeyEnter {
			// TODO: Process

			setInputStatus(false)
		} else if event.Key() == tcell.KeyEscape {
			setInputStatus(false)
		}

		return event
	}

	if event.Key() == tcell.KeyUp {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.HardDropPiece(0)
		activeGame.Unlock()
	} else if event.Key() == tcell.KeyDown {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, 0, -1)
		activeGame.Unlock()
	} else if event.Key() == tcell.KeyLeft {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, -1, 0)
		activeGame.Unlock()
	} else if event.Key() == tcell.KeyRight {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, 1, 0)
		activeGame.Unlock()
	} else if event.Rune() == 'z' || event.Rune() == 'Z' {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(0, 1, 1)
		activeGame.Unlock()
	} else if event.Rune() == 'x' || event.Rune() == 'X' {
		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(0, 1, 0)
		activeGame.Unlock()
	} else if event.Key() == tcell.KeyEnter {
		setInputStatus(!inputActive)
	} else if event.Key() == tcell.KeyEscape {
		done <- true
	}
	return event
}
