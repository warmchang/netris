package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"git.sr.ht/~tslocum/netris/pkg/event"
	"github.com/gdamore/tcell"
)

func handleKeypress(ev *tcell.EventKey) *tcell.EventKey {
	k := ev.Key()
	r := ev.Rune()

	if titleVisible {
		// TODO: During keybind change, record key, rune and modifier
		if titleScreen > 1 {
			switch k {
			case tcell.KeyEscape:
				titleScreen = 1
				titleSelectedButton = 0

				app.SetRoot(titleContainerGrid, true)
				updateTitle()
				return nil
			}

			return ev
		}

		switch k {
		case tcell.KeyEnter:
			if titleScreen == 1 {
				switch titleSelectedButton {
				case 0:
					resetPlayerSettingsForm()

					titleScreen = 2
					titleSelectedButton = 0

					app.SetRoot(playerSettingsContainerGrid, true)
					app.SetFocus(playerSettingsForm)
					app.Draw()
				case 1:
					resetGameSettingsForm()

					titleScreen = 3
					titleSelectedButton = 0

					app.SetRoot(gameSettingsContainerGrid, true)
					app.SetFocus(gameSettingsForm)
					app.Draw()
				case 2:
					titleScreen = 0

					updateTitle()
				}
			} else {
				if joinedGame {
					switch titleSelectedButton {
					case 0:
						setTitleVisible(false)
					case 1:
						titleScreen = 1
						titleSelectedButton = 0

						updateTitle()
					case 2:
						done <- true
					}
				} else {
					switch titleSelectedButton {
					case 0:
						selectMode <- event.ModePlayOnline
					case 1:
						selectMode <- event.ModePractice
					case 2:
						titleScreen = 1
						titleSelectedButton = 0

						updateTitle()
					}
				}
			}
		case tcell.KeyUp, tcell.KeyBacktab:
			previousTitleButton()
		case tcell.KeyDown, tcell.KeyTab:
			nextTitleButton()
		case tcell.KeyEscape:
			if titleScreen == 1 {
				titleScreen = 0
				titleSelectedButton = 0
			} else if joinedGame {
				setTitleVisible(false)
			} else {
				done <- true
			}
		default:
			switch r {
			case 'k', 'K':
				previousTitleButton()
			case 'j', 'J':
				nextTitleButton()
			}
		}

		updateTitle()

		return ev
	}

	if inputActive {
		if k == tcell.KeyEnter {
			msg := inputView.GetText()
			if msg != "" {
				if strings.HasPrefix(msg, "/cpu") {
					if profileCPU == nil {
						if len(msg) < 5 {
							logMessage("Profile name must be specified")
						} else {
							profileName := strings.TrimSpace(msg[5:])

							var err error
							profileCPU, err = os.Create(profileName)
							if err != nil {
								log.Fatal(err)
							}

							err = pprof.StartCPUProfile(profileCPU)
							if err != nil {
								log.Fatal(err)
							}

							logMessage(fmt.Sprintf("Started profiling CPU usage as %s", profileName))
						}
					} else {
						pprof.StopCPUProfile()
						profileCPU.Close()
						profileCPU = nil

						logMessage("Stopped profiling CPU usage")
					}
				} else {
					if activeGame != nil {
						activeGame.Event <- &event.MessageEvent{Message: msg}
					} else {
						logMessage("Message not sent - not currently connected to any game")
					}
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
		setTitleVisible(true)
	default:
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
	}

	return ev
}
