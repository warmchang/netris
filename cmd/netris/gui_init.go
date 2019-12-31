package main

import (
	"log"
	"unicode"

	"git.sr.ht/~tslocum/cview"
	"git.sr.ht/~tslocum/netris/pkg/event"
	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/gdamore/tcell"
)

func initGUI(skipTitle bool) (*cview.Application, error) {
	app = cview.NewApplication()

	app.SetAfterResizeFunc(handleResize)

	inputView = cview.NewInputField().
		SetText(DefaultStatusText).
		SetLabel("> ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	inputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !inputActive {
			return nil
		}

		return event
	})

	gameGrid = cview.NewGrid().
		SetBorders(false)

	mtx = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	mtx.SetDynamicColors(true)

	side = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	side.SetDynamicColors(true)

	buffer = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	buffer.SetDynamicColors(true)

	pad := cview.NewBox()

	recent = cview.NewTextView().
		SetScrollable(true).
		SetTextAlign(cview.AlignLeft).
		SetWrap(true).
		SetWordWrap(true)

	gameGrid.
		AddItem(pad, 0, 0, 4, 1, 0, 0, false).
		AddItem(pad, 0, 1, 1, 2, 0, 0, false).
		AddItem(mtx, 1, 1, 1, 1, 0, 0, false).
		AddItem(side, 1, 2, 1, 1, 0, 0, false).
		AddItem(buffer, 1, 3, 1, 1, 0, 0, false).
		AddItem(inputView, 2, 1, 1, 3, 0, 0, true).
		AddItem(recent, 3, 1, 1, 3, 0, 0, true)

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

	titleName = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	titleL = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	titleR = cview.NewTextView().
		SetScrollable(false).
		SetTextAlign(cview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	go handleTitle()

	buttonA = cview.NewButton("A")
	buttonLabelA = cview.NewTextView().SetTextAlign(cview.AlignCenter)

	buttonB = cview.NewButton("B")
	buttonLabelB = cview.NewTextView().SetTextAlign(cview.AlignCenter)

	buttonC = cview.NewButton("C")
	buttonLabelC = cview.NewTextView().SetTextAlign(cview.AlignCenter)

	titleNameGrid := cview.NewGrid().SetRows(3, 2).
		AddItem(titleName, 0, 0, 1, 1, 0, 0, false).
		AddItem(cview.NewTextView().SetText(SubTitle+game.Version), 1, 0, 1, 1, 0, 0, false)

	titleGrid = cview.NewGrid().
		SetRows(5, 3, 3, 3, 3, 3, 3).
		SetColumns(-1, 34, -1).
		AddItem(titleL, 0, 0, 8, 1, 0, 0, false).
		AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 8, 1, 0, 0, false).
		AddItem(buttonA, 1, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelA, 2, 1, 1, 1, 0, 0, false).
		AddItem(buttonB, 3, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelB, 4, 1, 1, 1, 0, 0, false).
		AddItem(buttonC, 5, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelC, 6, 1, 1, 1, 0, 0, false).
		AddItem(pad, 7, 1, 1, 1, 0, 0, false)

	gameListView = cview.NewTextView().SetDynamicColors(true)

	gameListButtonsGrid := cview.NewGrid().
		SetColumns(-1, 1, -1, 1, -1).
		AddItem(buttonA, 0, 0, 1, 1, 0, 0, false).
		AddItem(pad, 0, 1, 1, 1, 0, 0, false).
		AddItem(buttonB, 0, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 3, 1, 1, 0, 0, false).
		AddItem(buttonC, 0, 4, 1, 1, 0, 0, false)

	gameListHeader = cview.NewTextView().SetTextAlign(cview.AlignCenter)

	gameListGrid = cview.NewGrid().
		SetRows(5, 1, 14, 1, 3).
		SetColumns(-1, 34, -1).
		AddItem(titleL, 0, 0, 5, 1, 0, 0, false).
		AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 5, 1, 0, 0, false).
		AddItem(gameListHeader, 1, 1, 1, 1, 0, 0, true).
		AddItem(gameListView, 2, 1, 1, 1, 0, 0, true).
		AddItem(gameListButtonsGrid, 3, 1, 1, 1, 0, 0, true).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("\nRefresh: R\nPrevious: Shift+Tab - Next: Tab"), 4, 1, 1, 1, 0, 0, true)

	buttonCancel = cview.NewButton("Cancel")
	buttonStart = cview.NewButton("Start")

	newGameSubmitGrid := cview.NewGrid().
		SetColumns(-1, 10, 1, 10, -1).
		AddItem(pad, 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonCancel, 0, 1, 1, 1, 0, 0, false).
		AddItem(pad, 0, 2, 1, 1, 0, 0, false).
		AddItem(buttonStart, 0, 3, 1, 1, 0, 0, false).
		AddItem(pad, 0, 4, 1, 1, 0, 0, false)

	newGameNameInput = cview.NewInputField().SetText("netris")
	newGameMaxPlayersInput = cview.NewInputField().SetFieldWidth(3).SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return unicode.IsDigit(lastChar) && len(textToCheck) <= 3
	})
	newGameSpeedLimitInput = cview.NewInputField().SetFieldWidth(3).SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		return unicode.IsDigit(lastChar) && len(textToCheck) <= 3
	})

	resetNewGameInputs()

	newGameNameGrid := cview.NewGrid().
		AddItem(cview.NewTextView().SetText("Name"), 0, 0, 1, 1, 0, 0, false).
		AddItem(newGameNameInput, 0, 1, 1, 1, 0, 0, false)

	newGameMaxPlayersGrid := cview.NewGrid().
		AddItem(cview.NewTextView().SetText("Player Limit"), 0, 0, 1, 1, 0, 0, false).
		AddItem(newGameMaxPlayersInput, 0, 1, 1, 1, 0, 0, false)

	newGameSpeedLimitGrid := cview.NewGrid().
		AddItem(cview.NewTextView().SetText("Speed Limit"), 0, 0, 1, 1, 0, 0, false).
		AddItem(newGameSpeedLimitInput, 0, 1, 1, 1, 0, 0, false)

	newGameHeader := cview.NewTextView().
		SetTextAlign(cview.AlignCenter).
		SetWrap(false).
		SetWordWrap(false).SetText("New Game")

	newGameGrid = cview.NewGrid().
		SetRows(5, 2, 1, 1, 1, 1, 1, 1, 1, -1, 3).
		SetColumns(-1, 34, -1).
		AddItem(titleL, 0, 0, 11, 1, 0, 0, false).
		AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 11, 1, 0, 0, false).
		AddItem(newGameHeader, 1, 1, 1, 1, 0, 0, false).
		AddItem(newGameNameGrid, 2, 1, 1, 1, 0, 0, false).
		AddItem(pad, 3, 1, 1, 1, 0, 0, false).
		AddItem(newGameMaxPlayersGrid, 4, 1, 1, 1, 0, 0, false).
		AddItem(pad, 5, 1, 1, 1, 0, 0, false).
		AddItem(newGameSpeedLimitGrid, 6, 1, 1, 1, 0, 0, false).
		AddItem(pad, 7, 1, 1, 1, 0, 0, false).
		AddItem(newGameSubmitGrid, 8, 1, 1, 1, 0, 0, false).
		AddItem(pad, 9, 1, 1, 1, 0, 0, false).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("\nLimits set to zero are disabled\nPrevious: Shift+Tab - Next: Tab"), 10, 1, 1, 1, 0, 0, false)

	playerSettingsTitle := cview.NewTextView().
		SetTextAlign(cview.AlignCenter).
		SetWrap(false).
		SetWordWrap(false).SetText("Player Settings")

	playerSettingsForm = cview.NewForm().SetButtonsAlign(cview.AlignCenter)

	playerSettingsGrid = cview.NewGrid().
		SetRows(5, 2, -1, 1).
		SetColumns(-1, 34, -1).
		AddItem(titleL, 0, 0, 4, 1, 0, 0, false).
		AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 4, 1, 0, 0, false).
		AddItem(playerSettingsTitle, 1, 1, 1, 1, 0, 0, true).
		AddItem(playerSettingsForm, 2, 1, 1, 1, 0, 0, true).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("Previous: Shift+Tab - Next: Tab"), 3, 1, 1, 1, 0, 0, true)

	gameSettingsTitle := cview.NewTextView().
		SetTextAlign(cview.AlignCenter).
		SetWrap(false).
		SetWordWrap(false).SetText("Game Settings")

	buttonGhostPiece = cview.NewButton("Enabled")

	ghostPieceGrid := cview.NewGrid().SetColumns(19, -1).
		AddItem(cview.NewTextView().SetText("Ghost Piece"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonGhostPiece, 0, 1, 1, 1, 0, 0, false)

	buttonKeybindRotateCCW = cview.NewButton("Set")
	buttonKeybindRotateCW = cview.NewButton("Set")
	buttonKeybindMoveLeft = cview.NewButton("Set")
	buttonKeybindMoveRight = cview.NewButton("Set")
	buttonKeybindSoftDrop = cview.NewButton("Set")
	buttonKeybindHardDrop = cview.NewButton("Set")
	buttonKeybindCancel = cview.NewButton("Cancel")
	buttonKeybindSave = cview.NewButton("Save")

	rotateCCWGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Rotate CCW"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindRotateCCW, 0, 1, 1, 1, 0, 0, false)

	rotateCWGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Rotate CW"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindRotateCW, 0, 1, 1, 1, 0, 0, false)

	moveLeftGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Move Left"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindMoveLeft, 0, 1, 1, 1, 0, 0, false)

	moveRightGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Move Right"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindMoveRight, 0, 1, 1, 1, 0, 0, false)

	softDropGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Soft Drop"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindSoftDrop, 0, 1, 1, 1, 0, 0, false)

	hardDropGrid := cview.NewGrid().SetColumns(27, -1).
		AddItem(cview.NewTextView().SetText("Hard Drop"), 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindHardDrop, 0, 1, 1, 1, 0, 0, false)

	gameSettingsSubmitGrid := cview.NewGrid().
		SetColumns(-1, 10, 1, 10, -1).
		AddItem(pad, 0, 0, 1, 1, 0, 0, false).
		AddItem(buttonKeybindCancel, 0, 1, 1, 1, 0, 0, false).
		AddItem(pad, 0, 2, 1, 1, 0, 0, false).
		AddItem(buttonKeybindSave, 0, 3, 1, 1, 0, 0, false).
		AddItem(pad, 0, 4, 1, 1, 0, 0, false)

	gameSettingsGrid = cview.NewGrid().
		SetRows(5, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, -1).
		SetColumns(-1, 34, -1).
		AddItem(titleL, 0, 0, 18, 1, 0, 0, false).
		AddItem(titleNameGrid, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 18, 1, 0, 0, false).
		AddItem(gameSettingsTitle, 1, 1, 1, 1, 0, 0, false).
		AddItem(pad, 2, 1, 1, 1, 0, 0, false).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("Options"), 3, 1, 1, 1, 0, 0, false).
		AddItem(ghostPieceGrid, 4, 1, 1, 1, 0, 0, false).
		AddItem(ghostPieceGrid, 5, 1, 1, 1, 0, 0, false).
		AddItem(pad, 6, 1, 1, 1, 0, 0, false).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("Keybindings"), 7, 1, 1, 1, 0, 0, false).
		AddItem(pad, 8, 1, 1, 1, 0, 0, false).
		AddItem(rotateCCWGrid, 9, 1, 1, 1, 0, 0, false).
		AddItem(rotateCWGrid, 10, 1, 1, 1, 0, 0, false).
		AddItem(moveLeftGrid, 11, 1, 1, 1, 0, 0, false).
		AddItem(moveRightGrid, 12, 1, 1, 1, 0, 0, false).
		AddItem(softDropGrid, 13, 1, 1, 1, 0, 0, false).
		AddItem(hardDropGrid, 14, 1, 1, 1, 0, 0, false).
		AddItem(pad, 15, 1, 1, 1, 0, 0, false).
		AddItem(gameSettingsSubmitGrid, 16, 1, 1, 1, 0, 0, false).
		AddItem(cview.NewTextView().
			SetTextAlign(cview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("\nPrevious: Shift+Tab - Next: Tab"), 17, 1, 1, 1, 0, 0, false)

	titleContainerGrid = cview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false).
		AddItem(pad, 1, 0, 1, 1, 0, 0, false).
		AddItem(titleGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(pad, 1, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	gameListContainerGrid = cview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false).
		AddItem(pad, 1, 0, 1, 1, 0, 0, false).
		AddItem(gameListGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(pad, 1, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	newGameContainerGrid = cview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false).
		AddItem(pad, 1, 0, 1, 1, 0, 0, false).
		AddItem(newGameGrid, 1, 1, 1, 1, 0, 0, false).
		AddItem(pad, 1, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	playerSettingsContainerGrid = cview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false).
		AddItem(pad, 1, 0, 1, 1, 0, 0, false).
		AddItem(playerSettingsGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(pad, 1, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	gameSettingsContainerGrid = cview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false).
		AddItem(pad, 1, 0, 1, 1, 0, 0, false).
		AddItem(gameSettingsGrid, 1, 1, 1, 1, 0, 0, false).
		AddItem(pad, 1, 2, 1, 1, 0, 0, false).
		AddItem(pad, 0, 0, 1, 3, 0, 0, false)

	app = app.SetInputCapture(handleKeypress)

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
		{mino.Point{0, 0}, mino.BlockSolidRed},
		{mino.Point{0, 1}, mino.BlockSolidRed},
		{mino.Point{0, 2}, mino.BlockSolidRed},
		{mino.Point{0, 3}, mino.BlockSolidRed},
		{mino.Point{0, 4}, mino.BlockSolidRed},
		{mino.Point{1, 3}, mino.BlockSolidRed},
		{mino.Point{2, 2}, mino.BlockSolidRed},
		{mino.Point{3, 1}, mino.BlockSolidRed},
		{mino.Point{4, 0}, mino.BlockSolidRed},
		{mino.Point{4, 1}, mino.BlockSolidRed},
		{mino.Point{4, 2}, mino.BlockSolidRed},
		{mino.Point{4, 3}, mino.BlockSolidRed},
		{mino.Point{4, 4}, mino.BlockSolidRed},

		// E
		{mino.Point{7, 0}, mino.BlockSolidYellow},
		{mino.Point{7, 1}, mino.BlockSolidYellow},
		{mino.Point{7, 2}, mino.BlockSolidYellow},
		{mino.Point{7, 3}, mino.BlockSolidYellow},
		{mino.Point{7, 4}, mino.BlockSolidYellow},
		{mino.Point{8, 0}, mino.BlockSolidYellow},
		{mino.Point{9, 0}, mino.BlockSolidYellow},
		{mino.Point{8, 2}, mino.BlockSolidYellow},
		{mino.Point{9, 2}, mino.BlockSolidYellow},
		{mino.Point{8, 4}, mino.BlockSolidYellow},
		{mino.Point{9, 4}, mino.BlockSolidYellow},

		// T
		{mino.Point{12, 4}, mino.BlockSolidGreen},
		{mino.Point{13, 4}, mino.BlockSolidGreen},
		{mino.Point{14, 0}, mino.BlockSolidGreen},
		{mino.Point{14, 1}, mino.BlockSolidGreen},
		{mino.Point{14, 2}, mino.BlockSolidGreen},
		{mino.Point{14, 3}, mino.BlockSolidGreen},
		{mino.Point{14, 4}, mino.BlockSolidGreen},
		{mino.Point{15, 4}, mino.BlockSolidGreen},
		{mino.Point{16, 4}, mino.BlockSolidGreen},

		// R
		{mino.Point{19, 0}, mino.BlockSolidCyan},
		{mino.Point{19, 1}, mino.BlockSolidCyan},
		{mino.Point{19, 2}, mino.BlockSolidCyan},
		{mino.Point{19, 3}, mino.BlockSolidCyan},
		{mino.Point{19, 4}, mino.BlockSolidCyan},
		{mino.Point{20, 2}, mino.BlockSolidCyan},
		{mino.Point{20, 4}, mino.BlockSolidCyan},
		{mino.Point{21, 2}, mino.BlockSolidCyan},
		{mino.Point{21, 4}, mino.BlockSolidCyan},
		{mino.Point{22, 0}, mino.BlockSolidCyan},
		{mino.Point{22, 1}, mino.BlockSolidCyan},
		{mino.Point{22, 3}, mino.BlockSolidCyan},

		// I
		{mino.Point{25, 0}, mino.BlockSolidBlue},
		{mino.Point{25, 1}, mino.BlockSolidBlue},
		{mino.Point{25, 2}, mino.BlockSolidBlue},
		{mino.Point{25, 3}, mino.BlockSolidBlue},
		{mino.Point{25, 4}, mino.BlockSolidBlue},

		// S
		{mino.Point{28, 0}, mino.BlockSolidMagenta},
		{mino.Point{29, 0}, mino.BlockSolidMagenta},
		{mino.Point{30, 0}, mino.BlockSolidMagenta},
		{mino.Point{31, 1}, mino.BlockSolidMagenta},
		{mino.Point{29, 2}, mino.BlockSolidMagenta},
		{mino.Point{30, 2}, mino.BlockSolidMagenta},
		{mino.Point{28, 3}, mino.BlockSolidMagenta},
		{mino.Point{29, 4}, mino.BlockSolidMagenta},
		{mino.Point{30, 4}, mino.BlockSolidMagenta},
		{mino.Point{31, 4}, mino.BlockSolidMagenta},
	}

	for _, titleBlock := range titleBlocks {
		if !m.SetBlock(centerStart+titleBlock.X, titleBlock.Y, titleBlock.Block, false) {
			log.Fatalf("failed to set title block %s", titleBlock.Point)
		}
	}

	return m
}
