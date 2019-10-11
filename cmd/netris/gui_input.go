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
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.HardDropPiece(0)
		activeGame.Unlock()

		return nil
	} else if event.Key() == tcell.KeyDown {
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, 0, -1)
		activeGame.Unlock()

		return nil
	} else if event.Key() == tcell.KeyLeft {
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, -1, 0)
		activeGame.Unlock()

		return nil
	} else if event.Key() == tcell.KeyRight {
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, 1, 0)
		activeGame.Unlock()

		return nil
	} else if event.Rune() == 'z' || event.Rune() == 'Z' {
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(0, 1, 1)
		activeGame.Unlock()

		return nil
	} else if event.Rune() == 'x' || event.Rune() == 'X' {
		if activeGame == nil {
			return event
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(0, 1, 0)
		activeGame.Unlock()

		return nil
	} else if event.Key() == tcell.KeyEnter {
		setInputStatus(!inputActive)
	} else if event.Key() == tcell.KeyEscape {
		done <- true
	}

	return event
}
