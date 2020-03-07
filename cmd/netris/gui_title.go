package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"gitlab.com/tslocum/cbind"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"gitlab.com/tslocum/netris/pkg/event"
	"gitlab.com/tslocum/netris/pkg/game"
	"gitlab.com/tslocum/netris/pkg/mino"
)

const (
	SubTitle = " .rocketnine.space         v"
)

type screen int

const (
	screenTitle screen = iota
	screenSettings
	screenPlayerSettings
	screenGameSettings
	screenGames
	screenNewGame
)

var (
	titleVisible     bool
	currentScreen    screen
	currentSelection int
	drawTitle        = make(chan struct{}, game.CommandQueueSize)

	titleGrid          *cview.Grid
	titleContainerGrid *cview.Grid

	gameListSelected int

	newGameGrid            *cview.Grid
	newGameNameInput       *cview.InputField
	newGameMaxPlayersInput *cview.InputField
	newGameSpeedLimitInput *cview.InputField

	playerSettingsGrid          *cview.Grid
	playerSettingsContainerGrid *cview.Grid
	playerSettingsNameInput     *cview.InputField

	gameList              []*game.ListedGame
	gameListHeader        *cview.TextView
	gameListView          *cview.TextView
	gameListGrid          *cview.Grid
	gameListContainerGrid *cview.Grid
	newGameContainerGrid  *cview.Grid

	gameSettingsGrid          *cview.Grid
	gameSettingsContainerGrid *cview.Grid
	gameGrid                  *cview.Grid

	titleName *cview.TextView
	titleL    *cview.TextView
	titleR    *cview.TextView

	titleMatrixL = newTitleMatrixSide()
	titleMatrix  = newTitleMatrixName()
	titleMatrixR = newTitleMatrixSide()
	titlePiecesL []*mino.Piece
	titlePiecesR []*mino.Piece

	buttonA *cview.Button
	buttonB *cview.Button
	buttonC *cview.Button

	buttonLabelA *cview.TextView
	buttonLabelB *cview.TextView
	buttonLabelC *cview.TextView
)

func previousTitleButton() {
	if currentSelection == 0 {
		return
	}

	currentSelection--
}

func nextTitleButton() {
	maxButton := 2
	if currentScreen == screenGames {
		maxButton = 3
	} else if currentScreen == screenNewGame {
		maxButton = 4
	}
	if currentSelection >= maxButton {
		return
	}

	currentSelection++
}

func selectTitleButton() {
	if !titleVisible {
		return
	}

	switch currentScreen {
	case screenSettings:
		switch currentSelection {
		case 0:
			resetPlayerSettingsForm()

			currentScreen = screenPlayerSettings
			currentSelection = 0

			app.SetRoot(playerSettingsContainerGrid, true).SetFocus(playerSettingsNameInput)
		case 1:
			currentScreen = screenGameSettings
			currentSelection = 0

			drawGhostPieceUnsaved = drawGhostPiece

			app.SetRoot(gameSettingsContainerGrid, true)
			updateTitle()
		case 2:
			currentScreen = screenTitle
			currentSelection = 0

			updateTitle()
		}
	case screenPlayerSettings:
		if currentSelection == 0 { // Name input
			return
		} else if currentSelection == 2 { // Save
			nicknameDraft := playerSettingsNameInput.GetText()
			if nicknameDraft != "" && game.Nickname(nicknameDraft) != nickname {
				nickname = game.Nickname(nicknameDraft)

				if activeGame != nil {
					activeGame.Event <- &event.NicknameEvent{Nickname: nickname}
				}
			}
		}

		currentScreen = screenSettings
		currentSelection = 0

		app.SetRoot(titleContainerGrid, true)
		updateTitle()
	case screenGameSettings:
		if currentSelection == 0 {
			drawGhostPieceUnsaved = !drawGhostPieceUnsaved
			updateTitle()
			return
		} else if currentSelection == 7 || currentSelection == 8 {
			if currentSelection == 8 {
				drawGhostPiece = drawGhostPieceUnsaved

				for _, bind := range draftKeybindings {
					if bind.k == tcell.KeyRune {
						inputConfig.SetRune(bind.m, bind.r, actionHandlers[bind.a])
					} else {
						inputConfig.SetKey(bind.m, bind.k, actionHandlers[bind.a])
					}

					encoded, err := cbind.Encode(bind.m, bind.k, bind.r)
					if err == nil && encoded != "" {
						// Remove existing keybinds
						for existingBindAction, existingBinds := range config.Input {
							for i, existingBind := range existingBinds {
								if existingBind == encoded {
									config.Input[existingBindAction] = append(config.Input[existingBindAction][:i], config.Input[existingBindAction][i+1:]...)
									break
								}
							}
						}
						// Set keybind
						config.Input[bind.a] = append(config.Input[bind.a], encoded)
					}
				}
			}
			draftKeybindings = nil

			currentScreen = screenSettings
			currentSelection = 0

			app.SetRoot(titleContainerGrid, true)
			updateTitle()
			return
		}

		modal := cview.NewModal().SetText("Press desired key(s) to set keybinding or press Escape to cancel.").ClearButtons()
		app.SetRoot(modal, true)

		capturingKeybind = true
	case screenGames:
		if currentSelection == 0 {
			if gameListSelected >= 0 && gameListSelected < len(gameList) {
				joinGame <- gameList[gameListSelected].ID
			}
		} else if currentSelection == 1 {
			currentScreen = screenNewGame
			currentSelection = 0

			resetNewGameInputs()
			app.SetRoot(newGameContainerGrid, true).SetFocus(nil)
			updateTitle()
		} else if currentSelection == 2 {
			currentScreen = screenNewGame
			currentSelection = 0

			modal := cview.NewModal().SetText("Joining another server by IP via GUI is not yet implemented.\nPlease re-launch netris with the --connect argument instead.\n\nPress Escape to return.").ClearButtons()
			app.SetRoot(modal, true)
		} else if currentSelection == 3 {
			currentScreen = screenTitle
			currentSelection = 0

			app.SetRoot(titleContainerGrid, true)
			updateTitle()
		}
	case screenNewGame:
		if currentSelection == 3 {
			currentScreen = screenGames
			gameListSelected = 0
			currentSelection = 0
			app.SetRoot(gameListContainerGrid, true)
			renderGameList()
			updateTitle()
		} else if currentSelection == 4 {
			joinGame <- event.GameIDNewCustom
		}
	default: // Title screen 0
		if joinedGame {
			switch currentSelection {
			case 0:
				setTitleVisible(false)
			case 1:
				currentScreen = screenSettings
				currentSelection = 0

				updateTitle()
			case 2:
				done <- true
			}
		} else {
			switch currentSelection {
			case 0:
				currentScreen = screenGames
				currentSelection = 0
				gameListSelected = 0

				refreshGameList()
				renderGameList()

				app.SetRoot(gameListContainerGrid, true).SetFocus(nil)
				updateTitle()
			case 1:
				joinGame <- event.GameIDNewLocal
			case 2:
				currentScreen = screenSettings
				currentSelection = 0

				updateTitle()
			}
		}
	}
}

func setTitleVisible(visible bool) {
	if titleVisible == visible {
		return
	}

	titleVisible = visible

	if !titleVisible {
		app.SetRoot(gameGrid, true)

		app.SetFocus(nil)
	} else {
		currentScreen = screenTitle
		currentSelection = 0

		drawTitle <- struct{}{}

		app.SetRoot(titleContainerGrid, true)

		updateTitle()
	}
}

func updateTitle() {
	switch currentScreen {
	case screenSettings:
		buttonA.SetLabel("Player Settings")
		buttonLabelA.SetText("\nChange name")

		buttonB.SetLabel("Game Settings")
		buttonLabelB.SetText("\nChange keybindings")

		buttonC.SetLabel("Return")
		buttonLabelC.SetText("\nReturn to the last screen")
	case screenGames:
		buttonA.SetLabel("New Game")

		buttonB.SetLabel("Join by IP")

		buttonC.SetLabel("Return")
	default:
		if joinedGame {
			buttonA.SetLabel("Resume")
			buttonLabelA.SetText("\nResume game in progress")

			buttonB.SetLabel("Settings")
			buttonLabelB.SetText("\nPlayer name, keybindings, etc.")

			buttonC.SetLabel("Quit")
			buttonLabelC.SetText("\nQuit game")
		} else {
			buttonA.SetLabel("Play")
			buttonLabelA.SetText("\nPlay with others")

			buttonB.SetLabel("Practice")
			buttonLabelB.SetText("\nPlay alone")

			buttonC.SetLabel("Settings")
			buttonLabelC.SetText("\nPlayer name, keybindings, etc.")
		}
	}

	switch currentScreen {
	case screenPlayerSettings:
		switch currentSelection {
		case 1:
			app.SetFocus(playerSettingsCancel)
		case 2:
			app.SetFocus(playerSettingsSave)
		default:
			app.SetFocus(playerSettingsNameInput)
		}
		return
	case screenGameSettings:
		if drawGhostPieceUnsaved {
			buttonGhostPiece.SetLabel("Enabled")
		} else {
			buttonGhostPiece.SetLabel("Disabled")
		}

		switch currentSelection {
		case 0:
			app.SetFocus(buttonGhostPiece)
		case 1:
			app.SetFocus(buttonKeybindRotateCCW)
		case 2:
			app.SetFocus(buttonKeybindRotateCW)
		case 3:
			app.SetFocus(buttonKeybindMoveLeft)
		case 4:
			app.SetFocus(buttonKeybindMoveRight)
		case 5:
			app.SetFocus(buttonKeybindSoftDrop)
		case 6:
			app.SetFocus(buttonKeybindHardDrop)
		case 7:
			app.SetFocus(buttonKeybindCancel)
		case 8:
			app.SetFocus(buttonKeybindSave)
		}
		return
	case screenGames:
		switch currentSelection {
		case 2:
			app.SetFocus(buttonB)
		case 3:
			app.SetFocus(buttonC)
		case 1:
			app.SetFocus(buttonA)
		default:
			app.SetFocus(nil)
		}
		return
	case screenNewGame:
		switch currentSelection {
		case 1:
			app.SetFocus(newGameMaxPlayersInput)
		case 2:
			app.SetFocus(newGameSpeedLimitInput)
		case 3:
			app.SetFocus(buttonNewGameCancel)
		case 4:
			app.SetFocus(buttonNewGameStart)
		default:
			app.SetFocus(newGameNameInput)
		}
		return
	default:
		if currentScreen > 1 {
			return
		}
	}

	switch currentSelection {
	case 1:
		app.SetFocus(buttonB)
	case 2:
		app.SetFocus(buttonC)
	default:
		app.SetFocus(buttonA)
	}
}

func handleTitle() {
	var t *time.Ticker
	for {
		if t == nil {
			t = time.NewTicker(850 * time.Millisecond)
		} else {
			select {
			case <-t.C:
			case <-drawTitle:
				if t != nil {
					t.Stop()
				}

				t = time.NewTicker(850 * time.Millisecond)
			}
		}

		if !titleVisible {
			continue
		}

		titleMatrixL.ClearOverlay()

		for _, p := range titlePiecesL {
			p.Y -= 1
			if p.Y < -3 {
				p.Y = titleMatrixL.H + 2
			}
			if rand.Intn(4) == 0 {
				p.Mino = p.Rotate(1, 0)
				p.ApplyRotation(1, 0)
			}

			for _, m := range p.Mino {
				titleMatrixL.SetBlock(p.X+m.X, p.Y+m.Y, p.Solid, true)
			}
		}

		titleMatrixR.ClearOverlay()

		for _, p := range titlePiecesR {
			p.Y -= 1
			if p.Y < -3 {
				p.Y = titleMatrixL.H + 2
			}
			if rand.Intn(4) == 0 {
				p.Mino = p.Rotate(1, 0)
				p.ApplyRotation(1, 0)
			}

			for _, m := range p.Mino {
				if !titleMatrixR.ValidPoint(p.X+m.X, p.Y+m.Y) || titleMatrixR.Block(p.X+m.X, p.Y+m.Y) != mino.BlockNone {
					continue
				}

				titleMatrixR.SetBlock(p.X+m.X, p.Y+m.Y, p.Solid, true)
			}
		}

		app.QueueUpdateDraw(renderTitle)
	}
}

func renderTitle() {
	var newBlock mino.Block
	for i, b := range titleMatrix.M {
		switch b {
		case mino.BlockSolidZ:
			newBlock = mino.BlockSolidT
		case mino.BlockSolidO:
			newBlock = mino.BlockSolidZ
		case mino.BlockSolidS:
			newBlock = mino.BlockSolidO
		case mino.BlockSolidI:
			newBlock = mino.BlockSolidS
		case mino.BlockSolidJ:
			newBlock = mino.BlockSolidI
		case mino.BlockSolidT:
			newBlock = mino.BlockSolidJ
		default:
			continue
		}

		titleMatrix.M[i] = newBlock
	}

	renderLock.Lock()

	renderMatrixes([]*mino.Matrix{titleMatrix})
	titleName.Clear()
	titleName.Write(renderBuffer.Bytes())

	renderMatrixes([]*mino.Matrix{titleMatrixL})
	titleL.Clear()
	titleL.Write(renderBuffer.Bytes())

	renderMatrixes([]*mino.Matrix{titleMatrixR})
	titleR.Clear()
	titleR.Write(renderBuffer.Bytes())

	renderLock.Unlock()
}

func renderGameList() {
	w := 34

	gameListView.Clear()
	gameListView.Write([]byte("\n"))

	gameListView.Write([]byte(fmt.Sprintf("%-27s%s", "Game", "Players")))
	gameListView.Write([]byte("\n"))

	h := 10

	for i, g := range gameList {
		p := strconv.Itoa(g.Players)
		if g.MaxPlayers > 0 {
			p += "/" + strconv.Itoa(g.MaxPlayers)
		}

		if currentSelection == 0 && gameListSelected == i {
			gameListView.Write([]byte("[#000000:#FFFFFF]"))
		}
		gameListView.Write([]byte(fmt.Sprintf("%-27s%7s", g.Name, p)))
		if currentSelection == 0 && gameListSelected == i {
			gameListView.Write([]byte("[-:-]"))
		}
		gameListView.Write([]byte("\n"))

		h--
	}

	if h > 0 {
		for i := 0; i < h; i++ {
			for i := 0; i < w; i++ {
				gameListView.Write([]byte(" "))
			}
		}
	}
}

func refreshGameList() {
	app.QueueUpdateDraw(func() {
		gameListHeader.SetText("Finding games...")
	})

	go func() {
		ok := fetchGameList()
		app.QueueUpdateDraw(func() {
			if !ok {
				gameListHeader.SetText("Failed to connect to game server")
				return
			}

			var plural string
			if len(gameList) != 1 {
				plural = "s"
			}

			gameListHeader.SetText(fmt.Sprintf("Found %d game%s", len(gameList), plural))
		})
	}()
}

func fetchGameList() bool {
	s, err := game.Connect(connectAddress)
	if err != nil {
		return false
	}

	s.Write(&game.GameCommandListGames{})

	t := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-t.C:
			return false
		case e := <-s.In:
			if e.Command() == game.CommandListGames {
				if p, ok := e.(*game.GameCommandListGames); ok {
					gameList = p.Games
					if gameListSelected >= len(gameList) {
						gameListSelected = len(gameList) - 1
					}

					app.QueueUpdateDraw(renderGameList)

					s.Close()

					if !t.Stop() {
						<-t.C
					}

					return true
				}
			}
		}
	}
}

func resetNewGameInputs() {
	newGameNameInput.SetText("netris")
	newGameMaxPlayersInput.SetText("0")
	newGameSpeedLimitInput.SetText("0")
}

func selectTitleFunc(i int) func() {
	return func() {
		currentSelection = i
		selectTitleButton()
	}
}

func styleButton(button *cview.Button) {
	button.
		SetLabelColor(tcell.ColorWhite).
		SetBackgroundColorActivated(tcell.ColorWhite)
}

func styleInputField(inputField *cview.InputField) {
	inputField.
		SetFieldTextColor(tcell.ColorWhite)
}
