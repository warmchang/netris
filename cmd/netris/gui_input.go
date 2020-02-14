package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cbind"
	"gitlab.com/tslocum/netris/pkg/event"
	"gitlab.com/tslocum/netris/pkg/game"
)

type Keybinding struct {
	k tcell.Key
	r rune
	m tcell.ModMask

	a event.GameAction
}

var actionHandlers = map[event.GameAction]func(*tcell.EventKey) *tcell.EventKey{
	event.ActionRotateCCW: rotateCCW,
	event.ActionRotateCW:  rotateCW,
	event.ActionMoveLeft:  moveLeft,
	event.ActionMoveRight: moveRight,
	event.ActionSoftDrop:  softDrop,
	event.ActionHardDrop:  hardDrop,
}

var inputConfig = cbind.NewConfiguration()

var draftKeybindings []*Keybinding

func setKeyBinds() error {
	if len(config.Input) == 0 {
		setDefaultKeyBinds()
	}

	for a, keys := range config.Input {
		a = event.GameAction(strings.ToLower(string(a)))
		handler := actionHandlers[a]
		if handler == nil {
			return fmt.Errorf("failed to set keybind for %s: unknown action", a)
		}

		for _, k := range keys {
			mod, key, ch, err := cbind.Decode(k)
			if err != nil {
				return fmt.Errorf("failed to set keybind %s for %s: %s", k, a, err)
			}

			if key == tcell.KeyRune {
				inputConfig.SetRune(mod, ch, handler)
			} else {
				inputConfig.SetKey(mod, key, handler)
			}
		}
	}

	return nil
}

func setDefaultKeyBinds() {
	config.Input = map[event.GameAction][]string{
		event.ActionRotateCCW: {"z", "Z"},
		event.ActionRotateCW:  {"x", "X"},
		event.ActionMoveLeft:  {"Left", "h", "H"},
		event.ActionMoveRight: {"Right", "l", "L"},
		event.ActionSoftDrop:  {"Down", "j", "J"},
		event.ActionHardDrop:  {"Up", "k", "K"},
	}
}

func scrollMessages(direction int) {
	var scroll int
	if showLogLines > 3 {
		scroll = (showLogLines - 2) * direction
	} else {
		scroll = showLogLines * direction
	}

	r, _ := recent.GetScrollOffset()
	r += scroll
	if r < 0 {
		r = 0
	}
	recent.ScrollTo(r, 0)

	draw <- event.DrawAll
}

// Render functions called here don't need to be queued (Draw is called when nil is returned)
func handleKeypress(ev *tcell.EventKey) *tcell.EventKey {
	k := ev.Key()
	r := ev.Rune()

	if capturingKeybind {
		capturingKeybind = false
		if k == tcell.KeyEscape {
			draftKeybindings = nil

			app.SetRoot(gameSettingsContainerGrid, true)
			updateTitle()

			return nil
		}

		for i, bind := range draftKeybindings {
			if (bind.k != 0 && bind.k != k) || (bind.r != 0 && bind.r != r) || (bind.m != 0 && bind.m != ev.Modifiers()) {
				continue
			}

			draftKeybindings = append(draftKeybindings[:i], draftKeybindings[i+1:]...)
			break
		}

		var action event.GameAction
		switch currentSelection {
		case 1:
			action = event.ActionRotateCCW
		case 2:
			action = event.ActionRotateCW
		case 3:
			action = event.ActionMoveLeft
		case 4:
			action = event.ActionMoveRight
		case 5:
			action = event.ActionSoftDrop
		case 6:
			action = event.ActionHardDrop
		default:
			log.Fatal("setting keybind for unknown action")
		}

		draftKeybindings = append(draftKeybindings, &Keybinding{k: k, r: r, m: ev.Modifiers(), a: action})

		app.SetRoot(gameSettingsContainerGrid, true)
		updateTitle()
		return nil
	} else if titleVisible {
		if currentScreen > 1 {
			switch k {
			case tcell.KeyEscape:
				if currentScreen == screenNewGame {
					currentScreen = screenGames
					gameListSelected = 0
					currentSelection = 0
					app.SetRoot(gameListContainerGrid, true)
					renderGameList()
					updateTitle()
					return nil
				} else if currentScreen == screenGames {
					currentScreen = screenTitle
				} else {
					currentScreen = screenSettings
				}
				currentSelection = 0

				app.SetRoot(titleContainerGrid, true)
				updateTitle()
				return nil
			}

			if currentScreen == screenPlayerSettings {
				switch k {
				case tcell.KeyTab:
					currentSelection++
					if currentSelection > 2 {
						currentSelection = 2
					}

					updateTitle()
					return nil
				case tcell.KeyBacktab:
					currentSelection--
					if currentSelection < 0 {
						currentSelection = 0
					}

					updateTitle()
					return nil
				case tcell.KeyEnter:
					selectTitleButton()
					return nil
				}
			} else if currentScreen == screenGameSettings {
				switch k {
				case tcell.KeyTab:
					currentSelection++
					if currentSelection > 8 {
						currentSelection = 8
					}

					updateTitle()
					return nil
				case tcell.KeyBacktab:
					currentSelection--
					if currentSelection < 0 {
						currentSelection = 0
					}

					updateTitle()
					return nil
				case tcell.KeyEnter:
					selectTitleButton()
					return nil
				}
			} else if currentScreen == screenGames {
				switch k {
				case tcell.KeyUp:
					if currentSelection == 0 {
						if gameListSelected > 0 {
							gameListSelected--
						}
						renderGameList()
					}
					return nil
				case tcell.KeyBacktab:
					previousTitleButton()
					updateTitle()
					renderGameList()
					return nil
				case tcell.KeyDown:
					if currentSelection == 0 {
						if gameListSelected < len(gameList)-1 {
							gameListSelected++
						}
						renderGameList()
					}
					return nil
				case tcell.KeyTab:
					nextTitleButton()
					updateTitle()
					renderGameList()
					return nil
				case tcell.KeyEnter:
					selectTitleButton()
					return nil
				default:
					if currentSelection == 0 {
						switch r {
						case 'j', 'J':
							if gameListSelected < len(gameList)-1 {
								gameListSelected++
							}
							renderGameList()
							return nil
						case 'k', 'K':
							if gameListSelected > 0 {
								gameListSelected--
							}
							renderGameList()
							return nil
						case 'r', 'R':
							refreshGameList()
							return nil
						}
					}
				}
			} else if currentScreen == screenNewGame {
				switch k {
				case tcell.KeyBacktab:
					previousTitleButton()
					updateTitle()
					return nil
				case tcell.KeyTab:
					nextTitleButton()
					updateTitle()
					return nil
				case tcell.KeyEnter:
					selectTitleButton()
					return nil
				}
			}

			return ev
		}

		switch k {
		case tcell.KeyEnter:
			selectTitleButton()
			return nil
		case tcell.KeyUp, tcell.KeyBacktab:
			previousTitleButton()
			updateTitle()
			return nil
		case tcell.KeyDown, tcell.KeyTab:
			nextTitleButton()
			updateTitle()
			return nil
		case tcell.KeyEscape:
			if currentScreen == screenSettings {
				currentScreen = screenTitle
				currentSelection = 0
				updateTitle()
			} else if joinedGame {
				setTitleVisible(false)
			} else {
				done <- true
			}
			return nil
		default:
			switch r {
			case 'k', 'K':
				previousTitleButton()
				updateTitle()
				return nil
			case 'j', 'J':
				nextTitleButton()
				updateTitle()
				return nil
			}
		}

		return ev
	}

	if inputActive {
		switch k {
		case tcell.KeyEnter:
			defer setInputStatus(false)

			msg := inputView.GetText()
			if strings.TrimSpace(msg) == "" {
				return nil
			}

			msgl := strings.ToLower(msg)
			switch {
			case strings.HasPrefix(msgl, "/nick"):
				if activeGame != nil && len(msg) > 6 {
					var oldnick string
					activeGame.Lock()
					if p, ok := activeGame.Players[activeGame.LocalPlayer]; ok {
						oldnick = p.Name
						p.Name = game.Nickname(msg[6:])
					} else {
						return nil
					}
					activeGame.ProcessActionL(event.ActionNick)
					if p, ok := activeGame.Players[activeGame.LocalPlayer]; ok {
						p.Name = oldnick
					}
					activeGame.Unlock()
				}
			case strings.HasPrefix(msgl, "/ping"):
				if activeGame != nil {
					activeGame.ProcessAction(event.ActionPing)
				}
			case strings.HasPrefix(msgl, "/stats"):
				if activeGame != nil {
					activeGame.ProcessAction(event.ActionStats)
				}
			case strings.HasPrefix(msgl, "/version"):
				v := game.Version
				if v == "" {
					v = "unknown"
				}

				logMessage(fmt.Sprintf("netris version %s", v))
			case strings.HasPrefix(msgl, "/cpu"):
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
			default:
				if activeGame != nil {
					activeGame.Event <- &event.MessageEvent{Message: msg}
				} else {
					logMessage("Message not sent - not currently connected to any game")
				}
			}

			return nil
		case tcell.KeyPgUp:
			scrollMessages(-1)
			return nil
		case tcell.KeyPgDn:
			scrollMessages(1)
			return nil
		case tcell.KeyEscape:
			setInputStatus(false)
			return nil
		}

		return ev
	}

	switch k {
	case tcell.KeyEnter:
		setInputStatus(!inputActive)
		return nil
	case tcell.KeyTab:
		setShowDetails(!showDetails)
		return nil
	case tcell.KeyPgUp:
		scrollMessages(-1)
		return nil
	case tcell.KeyPgDn:
		scrollMessages(1)
		return nil
	case tcell.KeyEscape:
		setTitleVisible(true)
		return nil
	}

	return inputConfig.Capture(ev)
}

func rotateCCW(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionRotateCCW)
	return nil
}

func rotateCW(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionRotateCW)
	return nil
}

func moveLeft(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionMoveLeft)
	return nil
}

func moveRight(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionMoveRight)
	return nil
}

func softDrop(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionSoftDrop)
	return nil
}

func hardDrop(ev *tcell.EventKey) *tcell.EventKey {
	if activeGame == nil {
		return ev
	}

	activeGame.ProcessAction(event.ActionHardDrop)
	return nil
}
