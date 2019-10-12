package main

import (
	"git.sr.ht/~tslocum/netris/pkg/event"
	"github.com/gdamore/tcell"
)

func handleKeypress(ev *tcell.EventKey) *tcell.EventKey {
	k := ev.Key()
	r := ev.Rune()

	if inputActive {
		if k == tcell.KeyEnter {
			msg := inputView.GetText()
			if msg != "" {
				if activeGame != nil {
					activeGame.Event <- &event.MessageEvent{Message: msg}
				} else {
					logMessage("Message not sent - not currently connected to any game")
				}
			}

			setInputStatus(false)
		} else if k == tcell.KeyEscape {
			setInputStatus(false)
		}

		return ev
	}

	switch k {
	case tcell.KeyUp:
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.HardDropPiece()
		activeGame.Unlock()

		return nil
	case tcell.KeyDown:
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, -1)
		activeGame.Unlock()

		return nil
	case tcell.KeyLeft:
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(-1, 0)
		activeGame.Unlock()

		return nil
	case tcell.KeyRight:
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(1, 0)
		activeGame.Unlock()

		return nil
	case tcell.KeyEnter:
		setInputStatus(!inputActive)
	case tcell.KeyTab:
		setShowDetails(!showDetails)
	case tcell.KeyEscape:
		done <- true
	}

	switch r {
	case 'z', 'Z':
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(1, 1)
		activeGame.Unlock()

		return nil
	case 'x', 'X':

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.RotatePiece(1, 0)
		activeGame.Unlock()

		return nil
	case 'h', 'H':
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(-1, 0)
		activeGame.Unlock()

		return nil
	case 'j', 'J':
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(0, -1)
		activeGame.Unlock()

		return nil
	case 'k', 'K':
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.HardDropPiece()
		activeGame.Unlock()

		return nil
	case 'l', 'L':
		if activeGame == nil {
			return ev
		}

		activeGame.Lock()
		activeGame.Players[activeGame.LocalPlayer].Matrix.MovePiece(1, 0)
		activeGame.Unlock()

		return nil
	}

	return ev
}
