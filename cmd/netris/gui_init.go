package main

import (
	"log"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"gitlab.com/tslocum/netris/pkg/event"
	"gitlab.com/tslocum/netris/pkg/game"
	"gitlab.com/tslocum/netris/pkg/mino"
)

func initGUI(skipTitle bool) (*cview.Application, error) {
	cview.Styles.TitleColor = tcell.ColorDefault
	cview.Styles.BorderColor = tcell.ColorDefault
	cview.Styles.PrimaryTextColor = tcell.ColorDefault
	cview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	app = cview.NewApplication()
	app.EnableMouse(true)
	app.SetAfterResizeFunc(handleResize)

	inputView = cview.NewInputField()
	inputView.SetText(DefaultStatusText)
	inputView.SetLabel("> ")
	inputView.SetFieldWidth(0)
	inputView.SetFieldBackgroundColor(tcell.ColorDefault)
	inputView.SetFieldTextColor(tcell.ColorDefault)
	inputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !inputActive {
			return nil
		}

		return event
	})

	gameGrid = cview.NewGrid()
	gameGrid.SetBorders(false)

	mtx = cview.NewTextView()
	mtx.SetScrollable(false)
	mtx.SetTextAlign(cview.AlignLeft)
	mtx.SetWrap(false)
	mtx.SetWordWrap(false)

	mtx.SetDynamicColors(true)

	side = cview.NewTextView()
	side.SetScrollable(false)
	side.SetTextAlign(cview.AlignLeft)
	side.SetWrap(false)
	side.SetWordWrap(false)

	side.SetDynamicColors(true)

	buffer = cview.NewTextView()
	buffer.SetScrollable(false)
	buffer.SetTextAlign(cview.AlignLeft)
	buffer.SetWrap(false)
	buffer.SetWordWrap(false)

	buffer.SetDynamicColors(true)

	pad := cview.NewBox()

	recent = cview.NewTextView()
	recent.SetScrollable(true)
	recent.SetTextAlign(cview.AlignLeft)
	recent.SetWrap(true)
	recent.SetWordWrap(true)

	gameGrid.
		AddItem(pad, 0, 0, 4, 1, 0, 0, false)
	gameGrid.AddItem(pad, 0, 1, 1, 2, 0, 0, false)
	gameGrid.AddItem(mtx, 1, 1, 1, 1, 0, 0, false)
	gameGrid.AddItem(side, 1, 2, 1, 1, 0, 0, false)
	gameGrid.AddItem(buffer, 1, 3, 1, 1, 0, 0, false)
	gameGrid.AddItem(inputView, 2, 1, 1, 3, 0, 0, true)
	gameGrid.AddItem(recent, 3, 1, 1, 3, 0, 0, true)

	// Set up title screen

	titleVisible = !skipTitle

	minos, err := mino.Generate(4)
	if err != nil {
		log.Fatalf("failed to render title: failed to generate minos: %s", err)
	}

	var (
		piece      *mino.Piece
		addToRight bool
		i          int
		offset     int
	)
	for y := 0; y < 11; y++ {
		for x := 0; x < 4; x++ {
			if !addToRight {
				offset = 3
			} else {
				offset = 2
			}

			piece = mino.NewPiece(minos[i], mino.Point{(x * 5) + offset, (y * 5)})

			i++
			if i == len(minos) {
				i = 0
			}

			if addToRight {
				titlePiecesR = append(titlePiecesR, piece)
			} else {
				titlePiecesL = append(titlePiecesL, piece)
			}

			addToRight = !addToRight
		}
	}

	titleName = cview.NewTextView()
	titleName.SetScrollable(false)
	titleName.SetTextAlign(cview.AlignLeft)
	titleName.SetWrap(false)
	titleName.SetWordWrap(false)
	titleName.SetDynamicColors(true)

	titleL = cview.NewTextView()
	titleL.SetScrollable(false)
	titleL.SetTextAlign(cview.AlignLeft)
	titleL.SetWrap(false)
	titleL.SetWordWrap(false)
	titleL.SetDynamicColors(true)

	titleR = cview.NewTextView()
	titleR.SetScrollable(false)
	titleR.SetTextAlign(cview.AlignLeft)
	titleR.SetWrap(false)
	titleR.SetWordWrap(false)
	titleR.SetDynamicColors(true)

	go handleTitle()

	buttonA = cview.NewButton("A")
	buttonA.SetSelectedFunc(func() {
		currentSelection = 0
		if currentScreen == screenGames {
			currentSelection++
		}
		selectTitleButton()
	})
	styleButton(buttonA)
	buttonLabelA = cview.NewTextView()
	buttonLabelA.SetTextAlign(cview.AlignCenter)

	buttonB = cview.NewButton("B")
	buttonB.SetSelectedFunc(func() {
		currentSelection = 1
		if currentScreen == screenGames {
			currentSelection++
		}
		selectTitleButton()
	})
	styleButton(buttonB)
	buttonLabelB = cview.NewTextView()
	buttonLabelB.SetTextAlign(cview.AlignCenter)

	buttonC = cview.NewButton("C")
	buttonC.SetSelectedFunc(func() {
		currentSelection = 2
		if currentScreen == screenGames {
			currentSelection++
		}
		selectTitleButton()
	})
	styleButton(buttonC)
	buttonLabelC = cview.NewTextView()
	buttonLabelC.SetTextAlign(cview.AlignCenter)

	subTitle := cview.NewTextView()
	subTitle.SetText(SubTitle + game.Version)

	titleNameGrid := cview.NewGrid()
	titleNameGrid.SetRows(3, 2)
	titleNameGrid.AddItem(titleName, 0, 0, 1, 1, 0, 0, false)
	titleNameGrid.AddItem(subTitle, 1, 0, 1, 1, 0, 0, false)

	titleGrid = cview.NewGrid()
	titleGrid.SetRows(5, 3, 3, 3, 3, 3, 3)
	titleGrid.SetColumns(-1, 34, -1)
	titleGrid.AddItem(titleL, 0, 0, 8, 1, 0, 0, false)
	titleGrid.AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(titleR, 0, 2, 8, 1, 0, 0, false)
	titleGrid.AddItem(buttonA, 1, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(buttonLabelA, 2, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(buttonB, 3, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(buttonLabelB, 4, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(buttonC, 5, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(buttonLabelC, 6, 1, 1, 1, 0, 0, false)
	titleGrid.AddItem(pad, 7, 1, 1, 1, 0, 0, false)

	gameListView = cview.NewTextView()
	gameListView.SetDynamicColors(true)

	gameListButtonsGrid := cview.NewGrid()
	gameListButtonsGrid.SetColumns(-1, 1, -1, 1, -1)
	gameListButtonsGrid.AddItem(buttonA, 0, 0, 1, 1, 0, 0, false)
	gameListButtonsGrid.AddItem(pad, 0, 1, 1, 1, 0, 0, false)
	gameListButtonsGrid.AddItem(buttonB, 0, 2, 1, 1, 0, 0, false)
	gameListButtonsGrid.AddItem(pad, 0, 3, 1, 1, 0, 0, false)
	gameListButtonsGrid.AddItem(buttonC, 0, 4, 1, 1, 0, 0, false)

	gameListHeader = cview.NewTextView()
	gameListHeader.SetTextAlign(cview.AlignCenter)

	gameListHelp := cview.NewTextView()
	gameListHelp.SetTextAlign(cview.AlignCenter)
	gameListHelp.SetWrap(false)
	gameListHelp.SetWordWrap(false)
	gameListHelp.SetText("\nRefresh: R\nPrevious: Shift+Tab - Next: Tab")

	gameListGrid = cview.NewGrid()
	gameListGrid.SetRows(5, 1, 14, 1, 3)
	gameListGrid.SetColumns(-1, 34, -1)
	gameListGrid.AddItem(titleL, 0, 0, 5, 1, 0, 0, false)
	gameListGrid.AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false)
	gameListGrid.AddItem(titleR, 0, 2, 5, 1, 0, 0, false)
	gameListGrid.AddItem(gameListHeader, 1, 1, 1, 1, 0, 0, true)
	gameListGrid.AddItem(gameListView, 2, 1, 1, 1, 0, 0, true)
	gameListGrid.AddItem(gameListButtonsGrid, 3, 1, 1, 1, 0, 0, true)
	gameListGrid.AddItem(gameListHelp, 4, 1, 1, 1, 0, 0, true)

	buttonNewGameCancel = cview.NewButton("Cancel")
	buttonNewGameCancel.SetSelectedFunc(selectTitleFunc(3))
	buttonNewGameStart = cview.NewButton("Start")
	buttonNewGameStart.SetSelectedFunc(selectTitleFunc(4))

	styleButton(buttonNewGameCancel)
	styleButton(buttonNewGameStart)

	newGameSubmitGrid := cview.NewGrid()
	newGameSubmitGrid.SetColumns(-1, 10, 1, 10, -1)
	newGameSubmitGrid.AddItem(pad, 0, 0, 1, 1, 0, 0, false)
	newGameSubmitGrid.AddItem(buttonNewGameCancel, 0, 1, 1, 1, 0, 0, false)
	newGameSubmitGrid.AddItem(pad, 0, 2, 1, 1, 0, 0, false)
	newGameSubmitGrid.AddItem(buttonNewGameStart, 0, 3, 1, 1, 0, 0, false)
	newGameSubmitGrid.AddItem(pad, 0, 4, 1, 1, 0, 0, false)

	newGameNameInput = cview.NewInputField()
	newGameNameInput.SetText("netris")
	newGameMaxPlayersInput = cview.NewInputField()
	newGameMaxPlayersInput.SetFieldWidth(3)
	newGameMaxPlayersInput.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return unicode.IsDigit(lastChar) && len(textToCheck) <= 3
	})
	newGameSpeedLimitInput = cview.NewInputField()
	newGameSpeedLimitInput.SetFieldWidth(3)
	newGameSpeedLimitInput.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return unicode.IsDigit(lastChar) && len(textToCheck) <= 3
	})

	styleInputField(newGameNameInput)
	styleInputField(newGameMaxPlayersInput)
	styleInputField(newGameSpeedLimitInput)

	resetNewGameInputs()

	newGameNameLabel := cview.NewTextView()
	newGameNameLabel.SetText("Name")

	newGameNameGrid := cview.NewGrid()
	newGameNameGrid.AddItem(newGameNameLabel, 0, 0, 1, 1, 0, 0, false)
	newGameNameGrid.AddItem(newGameNameInput, 0, 1, 1, 1, 0, 0, false)

	newGameMaxPlayersLabel := cview.NewTextView()
	newGameMaxPlayersLabel.SetText("Player Limit")

	newGameMaxPlayersGrid := cview.NewGrid()
	newGameMaxPlayersGrid.AddItem(newGameMaxPlayersLabel, 0, 0, 1, 1, 0, 0, false)
	newGameMaxPlayersGrid.AddItem(newGameMaxPlayersInput, 0, 1, 1, 1, 0, 0, false)

	newGameSpeedLimitLabel := cview.NewTextView()
	newGameSpeedLimitLabel.SetText("Speed Limit")

	newGameSpeedLimitGrid := cview.NewGrid()
	newGameSpeedLimitGrid.AddItem(newGameSpeedLimitLabel, 0, 0, 1, 1, 0, 0, false)
	newGameSpeedLimitGrid.AddItem(newGameSpeedLimitInput, 0, 1, 1, 1, 0, 0, false)

	newGameHeader := cview.NewTextView()
	newGameHeader.SetTextAlign(cview.AlignCenter)
	newGameHeader.SetWrap(false)
	newGameHeader.SetWordWrap(false)
	newGameHeader.SetText("New Game")

	newGameHelp := cview.NewTextView()
	newGameHelp.SetTextAlign(cview.AlignCenter)
	newGameHelp.SetWrap(false)
	newGameHelp.SetWordWrap(false)
	newGameHelp.SetText("\nLimits set to zero are disabled\nPrevious: Shift+Tab - Next: Tab")

	newGameGrid = cview.NewGrid()
	newGameGrid.SetRows(5, 2, 1, 1, 1, 1, 1, 1, 1, -1, 3)
	newGameGrid.SetColumns(-1, 34, -1)
	newGameGrid.AddItem(titleL, 0, 0, 11, 1, 0, 0, false)
	newGameGrid.AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(titleR, 0, 2, 11, 1, 0, 0, false)
	newGameGrid.AddItem(newGameHeader, 1, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(newGameNameGrid, 2, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(pad, 3, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(newGameMaxPlayersGrid, 4, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(pad, 5, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(newGameSpeedLimitGrid, 6, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(pad, 7, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(newGameSubmitGrid, 8, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(pad, 9, 1, 1, 1, 0, 0, false)
	newGameGrid.AddItem(newGameHelp, 10, 1, 1, 1, 0, 0, false)

	playerSettingsTitle := cview.NewTextView()
	playerSettingsTitle.SetTextAlign(cview.AlignCenter)
	playerSettingsTitle.SetWrap(false)
	playerSettingsTitle.SetWordWrap(false)
	playerSettingsTitle.SetText("Player Settings")

	playerSettingsNameLabel := cview.NewTextView()
	playerSettingsNameLabel.SetText("Name")
	playerSettingsNameInput = cview.NewInputField()
	playerSettingsNameInput.SetFieldWidth(11)
	playerSettingsNameInput.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return len(textToCheck) <= 10
	})
	styleInputField(playerSettingsNameInput)

	playerSettingsNameGrid := cview.NewGrid()
	playerSettingsNameGrid.AddItem(playerSettingsNameLabel, 0, 0, 1, 1, 0, 0, false)
	playerSettingsNameGrid.AddItem(playerSettingsNameInput, 0, 1, 1, 1, 0, 0, false)

	playerSettingsCancel = cview.NewButton("Cancel")
	playerSettingsCancel.SetSelectedFunc(selectTitleFunc(1))
	playerSettingsSave = cview.NewButton("Save")
	playerSettingsSave.SetSelectedFunc(selectTitleFunc(2))

	styleButton(playerSettingsCancel)
	styleButton(playerSettingsSave)

	playerSettingsSubmitGrid := cview.NewGrid()
	playerSettingsSubmitGrid.SetColumns(-1, 10, 1, 10, -1)
	playerSettingsSubmitGrid.AddItem(pad, 0, 0, 1, 1, 0, 0, false)
	playerSettingsSubmitGrid.AddItem(playerSettingsCancel, 0, 1, 1, 1, 0, 0, false)
	playerSettingsSubmitGrid.AddItem(pad, 0, 2, 1, 1, 0, 0, false)
	playerSettingsSubmitGrid.AddItem(playerSettingsSave, 0, 3, 1, 1, 0, 0, false)
	playerSettingsSubmitGrid.AddItem(pad, 0, 4, 1, 1, 0, 0, false)

	playerSettingsHelp := cview.NewTextView()
	playerSettingsHelp.SetTextAlign(cview.AlignCenter)
	playerSettingsHelp.SetWrap(false)
	playerSettingsHelp.SetWordWrap(false)
	playerSettingsHelp.SetText("Previous: Shift+Tab - Next: Tab")

	playerSettingsGrid = cview.NewGrid()
	playerSettingsGrid.SetRows(5, 2, 1, 1, -1, 1, 1, 1)
	playerSettingsGrid.SetColumns(-1, 34, -1)
	playerSettingsGrid.AddItem(titleL, 0, 0, 8, 1, 0, 0, false)
	playerSettingsGrid.AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false)
	playerSettingsGrid.AddItem(titleR, 0, 2, 8, 1, 0, 0, false)
	playerSettingsGrid.AddItem(playerSettingsTitle, 1, 1, 1, 1, 0, 0, true)
	playerSettingsGrid.AddItem(pad, 2, 1, 1, 1, 0, 0, false)
	playerSettingsGrid.AddItem(playerSettingsNameGrid, 3, 1, 1, 1, 0, 0, true)
	playerSettingsGrid.AddItem(pad, 4, 1, 1, 1, 0, 0, false)
	playerSettingsGrid.AddItem(playerSettingsSubmitGrid, 5, 1, 1, 1, 0, 0, false)
	playerSettingsGrid.AddItem(pad, 6, 1, 1, 1, 0, 0, false)
	playerSettingsGrid.AddItem(playerSettingsHelp, 7, 1, 1, 1, 0, 0, true)

	gameSettingsTitle := cview.NewTextView()
	gameSettingsTitle.SetTextAlign(cview.AlignCenter)
	gameSettingsTitle.SetWrap(false)
	gameSettingsTitle.SetWordWrap(false)
	gameSettingsTitle.SetText("Game Settings")

	labelGhostPiece := cview.NewTextView()
	labelGhostPiece.SetText("Ghost Piece")

	buttonGhostPiece = cview.NewButton("Enabled")
	buttonGhostPiece.SetSelectedFunc(selectTitleFunc(0))
	styleButton(buttonGhostPiece)

	ghostPieceGrid := cview.NewGrid()
	ghostPieceGrid.SetColumns(19, -1)
	ghostPieceGrid.AddItem(labelGhostPiece, 0, 0, 1, 1, 0, 0, false)
	ghostPieceGrid.AddItem(buttonGhostPiece, 0, 1, 1, 1, 0, 0, false)

	labelKeybindRotateCCW := cview.NewTextView()
	labelKeybindRotateCCW.SetText("Rotate CCW")
	labelKeybindRotateCW := cview.NewTextView()
	labelKeybindRotateCW.SetText("Rotate CW")
	labelKeybindMoveLeft := cview.NewTextView()
	labelKeybindMoveLeft.SetText("Move Left")
	labelKeybindMoveRight := cview.NewTextView()
	labelKeybindMoveRight.SetText("Move Right")
	labelKeybindSoftDrop := cview.NewTextView()
	labelKeybindSoftDrop.SetText("Soft Drop")
	labelKeybindHardDrop := cview.NewTextView()
	labelKeybindHardDrop.SetText("Hard Drop")

	buttonKeybindRotateCCW = cview.NewButton("Set")
	buttonKeybindRotateCCW.SetSelectedFunc(selectTitleFunc(1))
	buttonKeybindRotateCW = cview.NewButton("Set")
	buttonKeybindRotateCW.SetSelectedFunc(selectTitleFunc(2))
	buttonKeybindMoveLeft = cview.NewButton("Set")
	buttonKeybindMoveLeft.SetSelectedFunc(selectTitleFunc(3))
	buttonKeybindMoveRight = cview.NewButton("Set")
	buttonKeybindMoveRight.SetSelectedFunc(selectTitleFunc(4))
	buttonKeybindSoftDrop = cview.NewButton("Set")
	buttonKeybindSoftDrop.SetSelectedFunc(selectTitleFunc(5))
	buttonKeybindHardDrop = cview.NewButton("Set")
	buttonKeybindHardDrop.SetSelectedFunc(selectTitleFunc(6))

	buttonKeybindCancel = cview.NewButton("Cancel")
	buttonKeybindCancel.SetSelectedFunc(selectTitleFunc(7))
	buttonKeybindSave = cview.NewButton("Save")
	buttonKeybindSave.SetSelectedFunc(selectTitleFunc(8))

	styleButton(buttonKeybindRotateCCW)
	styleButton(buttonKeybindRotateCW)
	styleButton(buttonKeybindMoveLeft)
	styleButton(buttonKeybindMoveRight)
	styleButton(buttonKeybindSoftDrop)
	styleButton(buttonKeybindHardDrop)
	styleButton(buttonKeybindCancel)
	styleButton(buttonKeybindSave)

	rotateCCWGrid := cview.NewGrid()
	rotateCCWGrid.SetColumns(27, -1)
	rotateCCWGrid.AddItem(labelKeybindRotateCCW, 0, 0, 1, 1, 0, 0, false)
	rotateCCWGrid.AddItem(buttonKeybindRotateCCW, 0, 1, 1, 1, 0, 0, false)

	rotateCWGrid := cview.NewGrid()
	rotateCWGrid.SetColumns(27, -1)
	rotateCWGrid.AddItem(labelKeybindRotateCW, 0, 0, 1, 1, 0, 0, false)
	rotateCWGrid.AddItem(buttonKeybindRotateCW, 0, 1, 1, 1, 0, 0, false)

	moveLeftGrid := cview.NewGrid()
	moveLeftGrid.SetColumns(27, -1)
	moveLeftGrid.AddItem(labelKeybindMoveLeft, 0, 0, 1, 1, 0, 0, false)
	moveLeftGrid.AddItem(buttonKeybindMoveLeft, 0, 1, 1, 1, 0, 0, false)

	moveRightGrid := cview.NewGrid()
	moveRightGrid.SetColumns(27, -1)
	moveRightGrid.AddItem(labelKeybindMoveRight, 0, 0, 1, 1, 0, 0, false)
	moveRightGrid.AddItem(buttonKeybindMoveRight, 0, 1, 1, 1, 0, 0, false)

	softDropGrid := cview.NewGrid()
	softDropGrid.SetColumns(27, -1)
	softDropGrid.AddItem(labelKeybindSoftDrop, 0, 0, 1, 1, 0, 0, false)
	softDropGrid.AddItem(buttonKeybindSoftDrop, 0, 1, 1, 1, 0, 0, false)

	hardDropGrid := cview.NewGrid()
	hardDropGrid.SetColumns(27, -1)
	hardDropGrid.AddItem(labelKeybindHardDrop, 0, 0, 1, 1, 0, 0, false)
	hardDropGrid.AddItem(buttonKeybindHardDrop, 0, 1, 1, 1, 0, 0, false)

	gameSettingsSubmitGrid := cview.NewGrid()
	gameSettingsSubmitGrid.SetColumns(-1, 10, 1, 10, -1)
	gameSettingsSubmitGrid.AddItem(pad, 0, 0, 1, 1, 0, 0, false)
	gameSettingsSubmitGrid.AddItem(buttonKeybindCancel, 0, 1, 1, 1, 0, 0, false)
	gameSettingsSubmitGrid.AddItem(pad, 0, 2, 1, 1, 0, 0, false)
	gameSettingsSubmitGrid.AddItem(buttonKeybindSave, 0, 3, 1, 1, 0, 0, false)
	gameSettingsSubmitGrid.AddItem(pad, 0, 4, 1, 1, 0, 0, false)

	gameSettingsOptionsTitle := cview.NewTextView()
	gameSettingsOptionsTitle.SetTextAlign(cview.AlignCenter)
	gameSettingsOptionsTitle.SetWrap(false)
	gameSettingsOptionsTitle.SetWordWrap(false)
	gameSettingsOptionsTitle.SetText("Options")

	gameSettingsKeybindsTitle := cview.NewTextView()
	gameSettingsKeybindsTitle.SetTextAlign(cview.AlignCenter)
	gameSettingsKeybindsTitle.SetWrap(false)
	gameSettingsKeybindsTitle.SetWordWrap(false)
	gameSettingsKeybindsTitle.SetText("Keybindings")

	gameSettingsHelp := cview.NewTextView()
	gameSettingsHelp.SetTextAlign(cview.AlignCenter)
	gameSettingsHelp.SetWrap(false)
	gameSettingsHelp.SetWordWrap(false)
	gameSettingsHelp.SetText("\nPrevious: Shift+Tab - Next: Tab")

	gameSettingsGrid = cview.NewGrid()
	gameSettingsGrid.SetRows(5, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, -1)
	gameSettingsGrid.SetColumns(-1, 34, -1)
	gameSettingsGrid.AddItem(titleL, 0, 0, 18, 1, 0, 0, false)
	gameSettingsGrid.AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(titleR, 0, 2, 18, 1, 0, 0, false)
	gameSettingsGrid.AddItem(gameSettingsTitle, 1, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(pad, 2, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(gameSettingsOptionsTitle, 3, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(ghostPieceGrid, 4, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(ghostPieceGrid, 5, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(pad, 6, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(gameSettingsKeybindsTitle, 7, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(pad, 8, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(rotateCCWGrid, 9, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(rotateCWGrid, 10, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(moveLeftGrid, 11, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(moveRightGrid, 12, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(softDropGrid, 13, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(hardDropGrid, 14, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(pad, 15, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(gameSettingsSubmitGrid, 16, 1, 1, 1, 0, 0, false)
	gameSettingsGrid.AddItem(gameSettingsHelp, 17, 1, 1, 1, 0, 0, false)

	titleContainerGrid = cview.NewGrid()
	titleContainerGrid.SetColumns(-1, 80, -1)
	titleContainerGrid.SetRows(-1, 24, -1)
	titleContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)
	titleContainerGrid.AddItem(pad, 1, 0, 1, 1, 0, 0, false)
	titleContainerGrid.AddItem(titleGrid, 1, 1, 1, 1, 0, 0, true)
	titleContainerGrid.AddItem(pad, 1, 2, 1, 1, 0, 0, false)
	titleContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	gameListContainerGrid = cview.NewGrid()
	gameListContainerGrid.SetColumns(-1, 80, -1)
	gameListContainerGrid.SetRows(-1, 24, -1)
	gameListContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)
	gameListContainerGrid.AddItem(pad, 1, 0, 1, 1, 0, 0, false)
	gameListContainerGrid.AddItem(gameListGrid, 1, 1, 1, 1, 0, 0, true)
	gameListContainerGrid.AddItem(pad, 1, 2, 1, 1, 0, 0, false)
	gameListContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	newGameContainerGrid = cview.NewGrid()
	newGameContainerGrid.SetColumns(-1, 80, -1)
	newGameContainerGrid.SetRows(-1, 24, -1)
	newGameContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)
	newGameContainerGrid.AddItem(pad, 1, 0, 1, 1, 0, 0, false)
	newGameContainerGrid.AddItem(newGameGrid, 1, 1, 1, 1, 0, 0, false)
	newGameContainerGrid.AddItem(pad, 1, 2, 1, 1, 0, 0, false)
	newGameContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	playerSettingsContainerGrid = cview.NewGrid()
	playerSettingsContainerGrid.SetColumns(-1, 80, -1)
	playerSettingsContainerGrid.SetRows(-1, 24, -1)
	playerSettingsContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)
	playerSettingsContainerGrid.AddItem(pad, 1, 0, 1, 1, 0, 0, false)
	playerSettingsContainerGrid.AddItem(playerSettingsGrid, 1, 1, 1, 1, 0, 0, true)
	playerSettingsContainerGrid.AddItem(pad, 1, 2, 1, 1, 0, 0, false)
	playerSettingsContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	gameSettingsContainerGrid = cview.NewGrid()
	gameSettingsContainerGrid.SetColumns(-1, 80, -1)
	gameSettingsContainerGrid.SetRows(-1, 24, -1)
	gameSettingsContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)
	gameSettingsContainerGrid.AddItem(pad, 1, 0, 1, 1, 0, 0, false)
	gameSettingsContainerGrid.AddItem(gameSettingsGrid, 1, 1, 1, 1, 0, 0, false)
	gameSettingsContainerGrid.AddItem(pad, 1, 2, 1, 1, 0, 0, false)
	gameSettingsContainerGrid.AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	app.SetInputCapture(handleKeypress)

	if !skipTitle {
		app.SetRoot(titleContainerGrid, true)

		updateTitle()
	} else {
		app.SetRoot(gameGrid, true)

		app.SetFocus(nil)
	}

	go handleDraw()

	return app, nil
}

func newTitleMatrixSide() *mino.Matrix {
	ev := make(chan interface{})
	go func() {
		for range ev {
		}
	}()

	draw := make(chan event.DrawObject)
	go func() {
		for range draw {
		}
	}()

	m := mino.NewMatrix(21, 48, 0, 1, ev, draw, mino.MatrixCustom)

	return m
}

func newTitleMatrixName() *mino.Matrix {
	ev := make(chan interface{})
	go func() {
		for range ev {
		}
	}()

	draw := make(chan event.DrawObject)
	go func() {
		for range draw {
		}
	}()

	m := mino.NewMatrix(36, 6, 0, 1, ev, draw, mino.MatrixCustom)

	centerStart := (m.W / 2) - 17

	var titleBlocks = []struct {
		mino.Point
		mino.Block
	}{
		// N
		{mino.Point{0, 0}, mino.BlockSolidZ},
		{mino.Point{0, 1}, mino.BlockSolidZ},
		{mino.Point{0, 2}, mino.BlockSolidZ},
		{mino.Point{0, 3}, mino.BlockSolidZ},
		{mino.Point{0, 4}, mino.BlockSolidZ},
		{mino.Point{1, 3}, mino.BlockSolidZ},
		{mino.Point{2, 2}, mino.BlockSolidZ},
		{mino.Point{3, 1}, mino.BlockSolidZ},
		{mino.Point{4, 0}, mino.BlockSolidZ},
		{mino.Point{4, 1}, mino.BlockSolidZ},
		{mino.Point{4, 2}, mino.BlockSolidZ},
		{mino.Point{4, 3}, mino.BlockSolidZ},
		{mino.Point{4, 4}, mino.BlockSolidZ},

		// E
		{mino.Point{7, 0}, mino.BlockSolidO},
		{mino.Point{7, 1}, mino.BlockSolidO},
		{mino.Point{7, 2}, mino.BlockSolidO},
		{mino.Point{7, 3}, mino.BlockSolidO},
		{mino.Point{7, 4}, mino.BlockSolidO},
		{mino.Point{8, 0}, mino.BlockSolidO},
		{mino.Point{9, 0}, mino.BlockSolidO},
		{mino.Point{8, 2}, mino.BlockSolidO},
		{mino.Point{9, 2}, mino.BlockSolidO},
		{mino.Point{8, 4}, mino.BlockSolidO},
		{mino.Point{9, 4}, mino.BlockSolidO},

		// T
		{mino.Point{12, 4}, mino.BlockSolidS},
		{mino.Point{13, 4}, mino.BlockSolidS},
		{mino.Point{14, 0}, mino.BlockSolidS},
		{mino.Point{14, 1}, mino.BlockSolidS},
		{mino.Point{14, 2}, mino.BlockSolidS},
		{mino.Point{14, 3}, mino.BlockSolidS},
		{mino.Point{14, 4}, mino.BlockSolidS},
		{mino.Point{15, 4}, mino.BlockSolidS},
		{mino.Point{16, 4}, mino.BlockSolidS},

		// R
		{mino.Point{19, 0}, mino.BlockSolidI},
		{mino.Point{19, 1}, mino.BlockSolidI},
		{mino.Point{19, 2}, mino.BlockSolidI},
		{mino.Point{19, 3}, mino.BlockSolidI},
		{mino.Point{19, 4}, mino.BlockSolidI},
		{mino.Point{20, 2}, mino.BlockSolidI},
		{mino.Point{20, 4}, mino.BlockSolidI},
		{mino.Point{21, 2}, mino.BlockSolidI},
		{mino.Point{21, 4}, mino.BlockSolidI},
		{mino.Point{22, 0}, mino.BlockSolidI},
		{mino.Point{22, 1}, mino.BlockSolidI},
		{mino.Point{22, 3}, mino.BlockSolidI},

		// I
		{mino.Point{25, 0}, mino.BlockSolidJ},
		{mino.Point{25, 1}, mino.BlockSolidJ},
		{mino.Point{25, 2}, mino.BlockSolidJ},
		{mino.Point{25, 3}, mino.BlockSolidJ},
		{mino.Point{25, 4}, mino.BlockSolidJ},

		// S
		{mino.Point{28, 0}, mino.BlockSolidT},
		{mino.Point{29, 0}, mino.BlockSolidT},
		{mino.Point{30, 0}, mino.BlockSolidT},
		{mino.Point{31, 1}, mino.BlockSolidT},
		{mino.Point{29, 2}, mino.BlockSolidT},
		{mino.Point{30, 2}, mino.BlockSolidT},
		{mino.Point{28, 3}, mino.BlockSolidT},
		{mino.Point{29, 4}, mino.BlockSolidT},
		{mino.Point{30, 4}, mino.BlockSolidT},
		{mino.Point{31, 4}, mino.BlockSolidT},
	}

	for _, titleBlock := range titleBlocks {
		if !m.SetBlock(centerStart+titleBlock.X, titleBlock.Y, titleBlock.Block, false) {
			log.Fatalf("failed to set title block %s", titleBlock.Point)
		}
	}

	return m
}
