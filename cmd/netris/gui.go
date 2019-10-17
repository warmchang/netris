package main

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
	"git.sr.ht/~tslocum/netris/pkg/game"

	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	closedGUI bool

	inputActive bool
	showDetails bool

	app                         *tview.Application
	titleGrid                   *tview.Grid
	titleContainerGrid          *tview.Grid
	playerSettingsForm          *tview.Form
	playerSettingsGrid          *tview.Grid
	playerSettingsContainerGrid *tview.Grid
	gameSettingsForm            *tview.Form
	gameSettingsGrid            *tview.Grid
	gameSettingsContainerGrid   *tview.Grid
	gameGrid                    *tview.Grid
	titleName                   *tview.TextView
	titleL                      *tview.TextView
	titleR                      *tview.TextView
	inputView                   *tview.InputField
	mtx                         *tview.TextView
	side                        *tview.TextView
	buffer                      *tview.TextView
	recent                      *tview.TextView

	joinedGame bool

	draw       = make(chan event.DrawObject, game.CommandQueueSize)
	selectMode = make(chan event.GameMode, game.CommandQueueSize)

	renderLock   = new(sync.Mutex)
	renderBuffer bytes.Buffer

	multiplayerMatrixSize int

	screenW, screenH       int
	newScreenW, newScreenH int

	nickname      = "Anonymous"
	nicknameDraft string

	inputHeight, mainHeight, newLogLines int
)

const DefaultStatusText = "Press Enter to chat, Z/X to rotate, arrow keys or HJKL to move/drop"

// TODO: Darken ghost color?
var renderBlock = map[mino.Block][]byte{
	mino.BlockNone:         []byte(" "),
	mino.BlockGhostBlue:    []byte("[#2864ff]▓[#ffffff]"), // 1a53ff
	mino.BlockSolidBlue:    []byte("[#2864ff]█[#ffffff]"),
	mino.BlockGhostCyan:    []byte("[#00eeee]▓[#ffffff]"),
	mino.BlockSolidCyan:    []byte("[#00eeee]█[#ffffff]"),
	mino.BlockGhostRed:     []byte("[#ee0000]▓[#ffffff]"),
	mino.BlockSolidRed:     []byte("[#ee0000]█[#ffffff]"),
	mino.BlockGhostYellow:  []byte("[#dddd00]▓[#ffffff]"),
	mino.BlockSolidYellow:  []byte("[#dddd00]█[#ffffff]"),
	mino.BlockGhostMagenta: []byte("[#c000cc]▓[#ffffff]"),
	mino.BlockSolidMagenta: []byte("[#c000cc]█[#ffffff]"),
	mino.BlockGhostGreen:   []byte("[#00e900]▓[#ffffff]"),
	mino.BlockSolidGreen:   []byte("[#00e900]█[#ffffff]"),
	mino.BlockGhostOrange:  []byte("[#ff7308]▓[#ffffff]"),
	mino.BlockSolidOrange:  []byte("[#ff7308]█[#ffffff]"),
	mino.BlockGarbage:      []byte("[#bbbbbb]█[#ffffff]"),
}

var (
	renderHLine    = []byte(string(tcell.RuneHLine))
	renderVLine    = []byte(string(tcell.RuneVLine))
	renderLLCorner = []byte(string(tcell.RuneLLCorner))
	renderLRCorner = []byte(string(tcell.RuneLRCorner))
)

func initGUI() (*tview.Application, error) {
	app = tview.NewApplication()

	app.SetBeforeDrawFunc(handleResize)

	inputView = tview.NewInputField().
		SetText(DefaultStatusText).
		SetLabel("").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault)

	inputView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if !inputActive {
			return nil
		}

		return event
	})

	gameGrid = tview.NewGrid().
		SetBorders(false).
		SetRows(2+(20*blockSize), -1)

	mtx = tview.NewTextView().
		SetScrollable(false).
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	mtx.SetDynamicColors(true)

	side = tview.NewTextView().
		SetScrollable(false).
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	side.SetDynamicColors(true)

	buffer = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false)

	buffer.SetDynamicColors(true)

	spacer := tview.NewBox()

	recent = tview.NewTextView().
		SetScrollable(false).
		SetTextAlign(tview.AlignLeft).
		SetWrap(true).
		SetWordWrap(true)

	gameGrid.SetColumns(1, 4+(10*blockSize), 10, -1).
		AddItem(spacer, 0, 0, 2, 1, 0, 0, false).
		AddItem(mtx, 0, 1, 1, 1, 0, 0, false).
		AddItem(side, 0, 2, 1, 1, 0, 0, false).
		AddItem(buffer, 0, 3, 1, 1, 0, 0, false).
		AddItem(inputView, 1, 1, 1, 3, 0, 0, true).
		AddItem(recent, 2, 1, 1, 3, 0, 0, true)

	// Set up title screen

	titleVisible = true

	minos, err := mino.Generate(4)
	if err != nil {
		log.Fatalf("failed to render title: failed to generate minos: %s", err)
	}

	var (
		piece      *mino.Piece
		addToRight bool
		i          int
	)
	for y := 0; y < 6; y++ {
		for x := 0; x < 4; x++ {
			piece = mino.NewPiece(minos[i], mino.Point{x * 5, (y * 5)})

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

	titleName = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	titleL = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	titleR = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetWrap(false).
		SetWordWrap(false).SetDynamicColors(true)

	go handleTitle()

	buttonA = tview.NewButton("A")
	buttonLabelA = tview.NewTextView().SetTextAlign(tview.AlignCenter)

	buttonB = tview.NewButton("B")
	buttonLabelB = tview.NewTextView().SetTextAlign(tview.AlignCenter)

	buttonC = tview.NewButton("C")
	buttonLabelC = tview.NewTextView().SetTextAlign(tview.AlignCenter)

	titleGrid = tview.NewGrid().
		SetRows(7, 3, 3, 3, 3, 3, 2).
		SetColumns(-1, 38, -1).
		AddItem(titleL, 0, 0, 7, 1, 0, 0, false).
		AddItem(titleName, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 7, 1, 0, 0, false).
		AddItem(buttonA, 1, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelA, 2, 1, 1, 1, 0, 0, false).
		AddItem(buttonB, 3, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelB, 4, 1, 1, 1, 0, 0, false).
		AddItem(buttonC, 5, 1, 1, 1, 0, 0, false).
		AddItem(buttonLabelC, 6, 1, 1, 1, 0, 0, false)

	playerSettingsTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetWordWrap(false).SetText("\nPlayer Settings")

	playerSettingsForm = tview.NewForm().SetButtonsAlign(tview.AlignCenter)

	playerSettingsGrid = tview.NewGrid().
		SetRows(7, 2, -1, 1).
		SetColumns(-1, 38, -1).
		AddItem(titleL, 0, 0, 3, 1, 0, 0, false).
		AddItem(titleName, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 3, 1, 0, 0, false).
		AddItem(playerSettingsTitle, 1, 1, 1, 1, 0, 0, true).
		AddItem(playerSettingsForm, 2, 1, 1, 1, 0, 0, true).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("Press Tab to move between fields"), 3, 1, 1, 1, 0, 0, true)

	gameSettingsTitle := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetWordWrap(false).SetText("\nGame Settings")

	gameSettingsForm = tview.NewForm().SetButtonsAlign(tview.AlignCenter)

	gameSettingsGrid = tview.NewGrid().
		SetRows(7, 2, -1, 1).
		SetColumns(-1, 38, -1).
		AddItem(titleL, 0, 0, 3, 1, 0, 0, false).
		AddItem(titleName, 0, 1, 1, 1, 0, 0, false).
		AddItem(titleR, 0, 2, 3, 1, 0, 0, false).
		AddItem(gameSettingsTitle, 1, 1, 1, 1, 0, 0, true).
		AddItem(gameSettingsForm, 2, 1, 1, 1, 0, 0, true).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetWrap(false).
			SetWordWrap(false).SetText("Press Tab to move between fields"), 3, 1, 1, 1, 0, 0, true)

	titleContainerGrid = tview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false).
		AddItem(tview.NewTextView(), 1, 0, 1, 1, 0, 0, false).
		AddItem(titleGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(tview.NewTextView(), 1, 2, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false)

	playerSettingsContainerGrid = tview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false).
		AddItem(tview.NewTextView(), 1, 0, 1, 1, 0, 0, false).
		AddItem(playerSettingsGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(tview.NewTextView(), 1, 2, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false)

	gameSettingsContainerGrid = tview.NewGrid().SetColumns(-1, 80, -1).SetRows(-1, 24, -1).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false).
		AddItem(tview.NewTextView(), 1, 0, 1, 1, 0, 0, false).
		AddItem(gameSettingsGrid, 1, 1, 1, 1, 0, 0, true).
		AddItem(tview.NewTextView(), 1, 2, 1, 1, 0, 0, false).
		AddItem(tview.NewTextView(), 0, 0, 1, 3, 0, 0, false)

	app = app.SetInputCapture(handleKeypress)

	app.SetRoot(titleContainerGrid, true)

	updateTitle()

	go handleDraw()

	return app, nil
}

func resetPlayerSettingsForm() {
	playerSettingsForm.Clear(true).AddInputField("Name", nickname, 0, nil, func(text string) {
		nicknameDraft = text
	}).AddButton("Cancel", func() {
		titleScreen = 1
		titleSelectedButton = 0

		app.SetRoot(titleContainerGrid, true)
		updateTitle()
	}).AddButton("Save", func() {
		if nicknameDraft != "" && game.Nickname(nicknameDraft) != nickname {
			nickname = game.Nickname(nicknameDraft)

			if activeGame != nil {
				activeGame.Event <- &event.NicknameEvent{Nickname: nickname}
			}
		}

		titleScreen = 1
		titleSelectedButton = 0

		app.SetRoot(titleContainerGrid, true)
		updateTitle()
	})
}

func resetGameSettingsForm() {
	gameSettingsForm.Clear(true).
		AddInputField("Custom", "", 0, nil, nil).
		AddInputField("Keybindings", "", 0, nil, nil).
		AddInputField("Are", "", 0, nil, nil).
		AddInputField("Coming", "", 0, nil, nil).
		AddInputField("Soon", "", 0, nil, nil).
		AddButton("Cancel", func() {
			titleScreen = 1
			titleSelectedButton = 0

			app.SetRoot(titleContainerGrid, true)
			updateTitle()
		}).AddButton("Save", func() {
		if nicknameDraft != "" && game.Nickname(nicknameDraft) != nickname {
			nickname = game.Nickname(nicknameDraft)

			if activeGame != nil {
				activeGame.Event <- &event.NicknameEvent{Nickname: nickname}
			}
		}

		titleScreen = 1
		titleSelectedButton = 0

		app.SetRoot(titleContainerGrid, true)
		updateTitle()
	})
}

func handleResize(screen tcell.Screen) bool {
	newScreenW, newScreenH = screen.Size()
	if newScreenW != screenW || newScreenH != screenH {
		screenW, screenH = newScreenW, newScreenH

		if !fixedBlockSize {
			if screenW >= 80 && screenH >= 44 {
				blockSize = 2
			} else {
				blockSize = 1
			}
		}

		multiplayerMatrixSize = (screenW - ((10 * blockSize) + 16)) / ((10 * blockSize) + 4)

		inputHeight = 1
		mainHeight = (20 * blockSize) + 2
		if screenH > mainHeight+5 {
			mainHeight += 2
			inputHeight++
		} else if screenH > mainHeight+2 {
			mainHeight++
		}

		newLogLines = (screenH - mainHeight) - inputHeight
		if newLogLines > 0 {
			showLogLines = newLogLines
		} else {
			showLogLines = 1
		}

		gameGrid.SetRows(mainHeight, inputHeight, -1).SetColumns(1, 4+(10*blockSize), 10, -1)

		logMutex.Lock()
		renderLogMessages = true
		logMutex.Unlock()
		draw <- event.DrawAll
		return true
	}

	return false
}

func drawAll() {
	if activeGame == nil {
		renderRecentMessages()

		return
	}

	renderPlayerMatrix()
	renderPreviewMatrix()
	renderMultiplayerMatrix()

	renderRecentMessages()
}

func drawPlayerMatrix() {
	renderPlayerMatrix()
	renderPreviewMatrix()

	renderRecentMessages()
}

func drawMultiplayerMatrixes() {
	renderMultiplayerMatrix()

	renderRecentMessages()
}

func handleDraw() {
	var o event.DrawObject
	for o = range draw {
		switch o {
		case event.DrawPlayerMatrix:
			app.QueueUpdateDraw(drawPlayerMatrix)
		case event.DrawMultiplayerMatrixes:
			app.QueueUpdateDraw(drawMultiplayerMatrixes)
		case event.DrawMessages:
			app.QueueUpdateDraw(renderRecentMessages)
		default:
		DRAW:
			for {
				select {
				case <-draw:
				default:
					break DRAW
				}
			}

			app.QueueUpdateDraw(drawAll)
		}
	}
}

func closeGUI() {
	if closedGUI {
		return
	}
	closedGUI = true

	app.Stop()
}

func setInputStatus(active bool) {
	if inputActive == active {
		return
	}

	inputActive = active

	if inputActive {
		inputView.SetText("")
		inputView.SetLabel("> ")
		app.SetFocus(inputView)
	} else {
		inputView.SetText(DefaultStatusText)
		inputView.SetLabel("")
		app.SetFocus(nil)
	}

	app.Draw()
}

func setShowDetails(active bool) {
	if showDetails == active {
		return
	}

	showDetails = active
	draw <- event.DrawAll
}

func renderPreviewMatrix() {
	g := activeGame
	if g == nil || len(g.Players) == 0 || g.Players[g.LocalPlayer].Matrix.Bag == nil {
		return
	}

	player := g.Players[g.LocalPlayer]
	m := g.Players[g.LocalPlayer].Matrix

	p := mino.NewPiece(m.Bag.Next(), mino.Point{0, 0})

	player.Preview.Clear()

	err := player.Preview.Add(p, p.Solid, mino.Point{0, 0}, false)
	if err != nil {
		log.Fatalf("failed to render preview matrix: %+v", err)
	}

	m.Lock()
	var (
		comboTime float64
		combo     int
	)
	if m.Combo > 0 && time.Since(m.ComboEnd) < 0 {
		comboTime = 1.0 + (float64(m.ComboEnd.Sub(time.Now())) / 1000000000)
		combo = m.Combo
	}
	m.Unlock()

	side.Clear()
	side.Write(renderMatrix(g.Players[g.LocalPlayer].Preview))
	m.Lock()
	renderLock.Lock()
	renderBuffer.Reset()
	if m.Speed < 100 {
		renderBuffer.WriteRune(' ')
	}
	renderBuffer.WriteString(strconv.Itoa(m.Speed))

	if blockSize > 1 {
		fmt.Fprint(side, fmt.Sprintf("\n\n\n\n\n Combo\n\n   %d\n\n\n\n\n Timer\n\n   %.0f\n\n\n\n\nPending\n\n   %d\n\n\n\n\n Speed\n\n  %s", combo, comboTime, m.PendingGarbage, renderBuffer.Bytes()))
	} else {
		fmt.Fprint(side, fmt.Sprintf("\n\n Combo\n\n   %d\n\n Timer\n\n   %.0f\n\nPending\n\n   %d\n\n Speed\n\n  %s", combo, comboTime, m.PendingGarbage, renderBuffer.Bytes()))
	}

	renderLock.Unlock()
	m.Unlock()
}

func renderPlayerMatrix() {
	g := activeGame
	if g == nil || len(g.Players) == 0 {
		return
	}

	mtx.Clear()
	mtx.Write(renderMatrix(g.Players[g.LocalPlayer].Matrix))
}

func renderMultiplayerMatrix() {
	g := activeGame
	if g == nil {
		return
	}

	g.Lock()

	if g.LocalPlayer == game.PlayerUnknown || len(g.Players) <= 1 {
		buffer.Clear()
		g.Unlock()
		return
	}

	var (
		playerIDs = make([]int, len(g.Players)-1)
		i         int
	)
	for playerID := range g.Players {
		if playerID == g.LocalPlayer {
			continue
		}

		playerIDs[i] = playerID
		i++
	}
	sort.Ints(playerIDs)

	i = 0
	var matrixes []*mino.Matrix
	for _, playerID := range playerIDs {
		if g.Players[playerID] == nil {
			continue
		}

		i++
		matrixes = append(matrixes, g.Players[playerID].Matrix)

		if i == multiplayerMatrixSize {
			break
		}
	}

	g.Unlock()

	buffer.Clear()
	buffer.Write(renderMatrixes(matrixes))
}

func renderMatrix(m *mino.Matrix) []byte {
	if m == nil {
		return nil
	}

	renderLock.Lock()
	defer renderLock.Unlock()

	m.Lock()
	defer m.Unlock()

	renderBuffer.Reset()

	m.DrawPiecesL()

	bs := blockSize
	if m.Type == mino.MatrixPreview {
		// Draw preview matrix at block size 2 max

		if bs > 2 {
			bs = 2
		}
		if bs > 1 {
			renderBuffer.WriteRune('\n')
		}
	} else if m.Type == mino.MatrixCustom {
		bs = 1
	}

	for y := m.H - 1; y >= 0; y-- {
		for j := 0; j < bs; j++ {
			if m.Type == mino.MatrixStandard {
				renderBuffer.Write(renderVLine)
			} else {
				iPieceNext := m.Bag != nil && m.Bag.Next().String() == mino.TetrominoI
				if bs == 1 {
					renderBuffer.WriteRune(' ')
					renderBuffer.WriteRune(' ')
				} else if !iPieceNext {
					renderBuffer.WriteRune(' ')
				}
			}
			for x := 0; x < m.W; x++ {
				for k := 0; k < bs; k++ {
					renderBuffer.Write(renderBlock[m.Block(x, y)])
				}
			}

			if m.Type == mino.MatrixStandard {
				renderBuffer.Write(renderVLine)
			}

			if y != 0 || m.Type == mino.MatrixStandard {
				renderBuffer.WriteRune('\n')
			}
		}
	}

	if m.Type != mino.MatrixStandard {
		return renderBuffer.Bytes()
	}

	renderBuffer.Write(renderLLCorner)
	for x := 0; x < m.W*bs; x++ {
		renderBuffer.Write(renderHLine)
	}
	renderBuffer.Write(renderLRCorner)

	renderBuffer.WriteRune('\n')
	renderPlayerDetails(m, bs)

	return renderBuffer.Bytes()
}

func renderPlayerDetails(m *mino.Matrix, bs int) {
	var buf string
	if !showDetails {
		buf = m.PlayerName
	} else {
		if blockSize == 1 {
			buf = fmt.Sprintf("%d/%d @ %d", m.GarbageSent, m.GarbageReceived, m.Speed)
		} else {
			buf = fmt.Sprintf("%d / %d  @  %d", m.GarbageSent, m.GarbageReceived, m.Speed)
		}
	}
	if len(buf) > m.W*bs {
		buf = buf[:m.W*bs]
	}

	padBuf := ((m.W*bs - len(buf)) / 2) + 1
	for i := 0; i < padBuf; i++ {
		renderBuffer.WriteRune(' ')
	}
	renderBuffer.WriteString(buf)
	padBuf = m.W*bs + 2 - len(buf) - padBuf
	for i := 0; i < padBuf; i++ {
		renderBuffer.WriteRune(' ')
	}
}

func renderMatrixes(mx []*mino.Matrix) []byte {
	if mx == nil {
		return nil
	}

	renderLock.Lock()
	defer renderLock.Unlock()

	for i := range mx {
		mx[i].Lock()
		mx[i].DrawPiecesL()
	}

	renderBuffer.Reset()

	div := "  "

	height := mx[0].H

	for y := height - 1; y >= 0; y-- {
		for j := 0; j < blockSize; j++ {
			for i := range mx {
				m := mx[i]

				if i > 0 {
					renderBuffer.WriteString(div)
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				}

				for x := 0; x < m.W; x++ {
					for j := 0; j < blockSize; j++ {
						renderBuffer.Write(renderBlock[m.Block(x, y)])
					}
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				}
			}

			renderBuffer.WriteRune('\n')
		}
	}

	for i := range mx {
		if i > 0 {
			renderBuffer.WriteString(div)
		}

		renderBuffer.Write(renderLLCorner)
		for x := 0; x < mx[i].W*blockSize; x++ {
			renderBuffer.Write(renderHLine)
		}
		renderBuffer.Write(renderLRCorner)
	}

	renderBuffer.WriteRune('\n')

	for i, m := range mx {
		if i > 0 {
			renderBuffer.WriteString(div)
		}

		renderPlayerDetails(m, blockSize)
	}

	for i := range mx {
		mx[i].Unlock()
	}

	return renderBuffer.Bytes()
}

func logMessage(message string) {
	logMutex.Lock()
	logMessages = append(logMessages, time.Now().Format(LogTimeFormat)+" "+message)
	renderLogMessages = true
	logMutex.Unlock()
}

func renderRecentMessages() {
	logMutex.Lock()
	if !renderLogMessages {
		logMutex.Unlock()
		return
	}

	l := len(logMessages)
	ls := l - showLogLines
	if ls < 0 {
		ls = 0
	}
	recent.SetText(strings.Join(logMessages[ls:l], "\n"))

	renderLogMessages = false
	logMutex.Unlock()
}
