package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/tslocum/tview"
)

const (
	SubTitle = " .rocketnine.space         v"
)

var (
	titleVisible               bool
	titleScreen                int
	titleSelectedButton        int
	gameSettingsSelectedButton int
	drawTitle                  = make(chan struct{}, game.CommandQueueSize)

	titleGrid          *tview.Grid
	titleContainerGrid *tview.Grid

	gameListSelected int

	newGameGrid            *tview.Grid
	newGameNameInput       *tview.InputField
	newGameMaxPlayersInput *tview.InputField
	newGameSpeedLimitInput *tview.InputField

	playerSettingsForm          *tview.Form
	playerSettingsGrid          *tview.Grid
	playerSettingsContainerGrid *tview.Grid

	gameList              []*game.ListedGame
	gameListHeader        *tview.TextView
	gameListView          *tview.TextView
	gameListGrid          *tview.Grid
	gameListContainerGrid *tview.Grid
	newGameContainerGrid  *tview.Grid

	gameSettingsGrid          *tview.Grid
	gameSettingsContainerGrid *tview.Grid
	gameGrid                  *tview.Grid

	titleName *tview.TextView
	titleL    *tview.TextView
	titleR    *tview.TextView

	titleMatrixL = newTitleMatrixSide()
	titleMatrix  = newTitleMatrixName()
	titleMatrixR = newTitleMatrixSide()
	titlePiecesL []*mino.Piece
	titlePiecesR []*mino.Piece

	buttonA *tview.Button
	buttonB *tview.Button
	buttonC *tview.Button

	buttonLabelA *tview.TextView
	buttonLabelB *tview.TextView
	buttonLabelC *tview.TextView
)

func previousTitleButton() {
	if titleSelectedButton == 0 {
		return
	}

	titleSelectedButton--
}

func nextTitleButton() {
	maxButton := 2
	if titleScreen == 4 {
		maxButton = 3
	} else if titleScreen == 5 {
		maxButton = 4
	}
	if titleSelectedButton >= maxButton {
		return
	}

	titleSelectedButton++
}

func updateGameSettings() {
	switch gameSettingsSelectedButton {
	case 0:
		app.SetFocus(buttonKeybindRotateCCW)
	case 1:
		app.SetFocus(buttonKeybindRotateCW)
	case 2:
		app.SetFocus(buttonKeybindMoveLeft)
	case 3:
		app.SetFocus(buttonKeybindMoveRight)
	case 4:
		app.SetFocus(buttonKeybindSoftDrop)
	case 5:
		app.SetFocus(buttonKeybindHardDrop)
	case 6:
		app.SetFocus(buttonKeybindCancel)
	case 7:
		app.SetFocus(buttonKeybindSave)
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
		titleScreen = 0
		titleSelectedButton = 0

		drawTitle <- struct{}{}

		app.SetRoot(titleContainerGrid, true)

		updateTitle()
	}
}

func updateTitle() {
	if titleScreen == 1 {
		buttonA.SetLabel("Player Settings")
		buttonLabelA.SetText("\nChange name")

		buttonB.SetLabel("Game Settings")
		buttonLabelB.SetText("\nChange keybindings")

		buttonC.SetLabel("Return")
		buttonLabelC.SetText("\nReturn to the last screen")
	} else if titleScreen == 4 {
		buttonA.SetLabel("New Game")

		buttonB.SetLabel("Join by IP")

		buttonC.SetLabel("Return")
	} else {
		if joinedGame {
			buttonA.SetLabel("Resume")
			buttonLabelA.SetText("\nResume game in progress")

			buttonB.SetLabel("Settings")
			buttonLabelB.SetText("\nChange player name, keybindings, etc")

			buttonC.SetLabel("Quit")
			buttonLabelC.SetText("\nQuit game")
		} else {
			buttonA.SetLabel("Play")
			buttonLabelA.SetText("\nPlay with others online")

			buttonB.SetLabel("Practice")
			buttonLabelB.SetText("\nPlay by yourself")

			buttonC.SetLabel("Settings")
			buttonLabelC.SetText("\nPlayer name, keybindings, etc.")
		}
	}

	if titleScreen == 4 {
		switch titleSelectedButton {
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
	} else if titleScreen == 5 {
		switch titleSelectedButton {
		case 1:
			app.SetFocus(newGameMaxPlayersInput)
		case 2:
			app.SetFocus(newGameSpeedLimitInput)
		case 3:
			app.SetFocus(buttonCancel)
		case 4:
			app.SetFocus(buttonStart)
		default:
			app.SetFocus(newGameNameInput)
		}

		return
	} else if titleScreen > 1 {
		return
	}

	switch titleSelectedButton {
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
				if titleMatrixR.Block(p.X+m.X, p.Y+m.Y) != mino.BlockNone {
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
		case mino.BlockSolidRed:
			newBlock = mino.BlockSolidMagenta
		case mino.BlockSolidYellow:
			newBlock = mino.BlockSolidRed
		case mino.BlockSolidGreen:
			newBlock = mino.BlockSolidYellow
		case mino.BlockSolidCyan:
			newBlock = mino.BlockSolidGreen
		case mino.BlockSolidBlue:
			newBlock = mino.BlockSolidCyan
		case mino.BlockSolidMagenta:
			newBlock = mino.BlockSolidBlue
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
	w := 32

	gameListView.Clear()
	gameListView.Write(renderULCorner)
	for i := 0; i < w; i++ {
		gameListView.Write(renderHLine)
	}
	gameListView.Write(renderURCorner)
	gameListView.Write([]byte("\n"))

	gameListView.Write(renderVLine)
	gameListView.Write([]byte(fmt.Sprintf("%-25s%s", "Game", "Players")))
	gameListView.Write(renderVLine)
	gameListView.Write([]byte("\n"))

	gameListView.Write(renderLTee)
	for i := 0; i < w; i++ {
		gameListView.Write(renderHLine)
	}
	gameListView.Write(renderRTee)
	gameListView.Write([]byte("\n"))

	h := 10

	for i, g := range gameList {
		p := strconv.Itoa(g.Players)
		if g.MaxPlayers > 0 {
			p += "/" + strconv.Itoa(g.MaxPlayers)
		}

		gameListView.Write(renderVLine)
		if titleSelectedButton == 0 && gameListSelected == i {
			gameListView.Write([]byte("[#000000:#FFFFFF]"))
		}
		gameListView.Write([]byte(fmt.Sprintf("%-25s%7s", g.Name, p)))
		if titleSelectedButton == 0 && gameListSelected == i {
			gameListView.Write([]byte("[-:-]"))
		}
		gameListView.Write(renderVLine)
		gameListView.Write([]byte("\n"))

		h--
	}

	if h > 0 {
		for i := 0; i < h; i++ {
			gameListView.Write(renderVLine)
			for i := 0; i < w; i++ {
				gameListView.Write([]byte(" "))
			}
			gameListView.Write(renderVLine)
		}
	}

	gameListView.Write(renderLLCorner)
	for i := 0; i < w; i++ {
		gameListView.Write(renderHLine)
	}
	gameListView.Write(renderLRCorner)
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
