package main

import (
	"git.sr.ht/~tslocum/netris/pkg/event"
	"github.com/gdamore/tcell"
)

func handleKeypress(ev *tcell.EventKey) *tcell.EventKey {
	if inputActive {
		if ev.Key() == tcell.KeyEnter {
			msg := inputView.GetText()
			if msg != "" {
				if activeGame != nil {
					activeGame.Event <- &event.MessageEvent{Message: msg}
				} else {
					logMessage("Message not sent - not currently connected to any game")
				}
			}

			setInputStatus(false)
		} else if ev.Key() == tcell.KeyEscape {
			setInputStatus(false)
		}

		return ev
	}

	if ev.Key() == tcell.KeyUp {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.HardDropPiece()
		activeGame.Unlock()

		return nil
	} else if ev.Key() == tcell.KeyDown {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, -1)
		activeGame.Unlock()

		return nil
	} else if ev.Key() == tcell.KeyLeft {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(-1, 0)
		activeGame.Unlock()

		return nil
	} else if ev.Key() == tcell.KeyRight {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(1, 0)
		activeGame.Unlock()

		return nil
	} else if ev.Rune() == 'z' || ev.Rune() == 'Z' {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(1, 1)
		activeGame.Unlock()

		return nil
	} else if ev.Rune() == 'x' || ev.Rune() == 'X' {
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(1, 0)
		activeGame.Unlock()

		return nil
	} else if ev.Key() == tcell.KeyEnter {
		setInputStatus(!inputActive)
	} else if ev.Key() == tcell.KeyEscape {
		done <- true
	}

	return ev
}
