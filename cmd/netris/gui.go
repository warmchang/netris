package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/event"
	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/gdamore/tcell"
	"github.com/tslocum/tview"
)

var (
	closedGUI bool

	inputActive      bool
	capturingKeybind bool
	showDetails      bool

	app       *tview.Application
	inputView *tview.InputField
	mtx       *tview.TextView
	side      *tview.TextView
	buffer    *tview.TextView
	recent    *tview.TextView

	joinedGame bool

	draw     = make(chan event.DrawObject, game.CommandQueueSize)
	joinGame = make(chan int, game.CommandQueueSize)

	renderLock   = new(sync.Mutex)
	renderBuffer bytes.Buffer

	multiplayerMatrixSize int
	extraScreenPadding    int

	screenW, screenH       int
	newScreenW, newScreenH int

	nickname      = "Anonymous"
	nicknameDraft string

	inputHeight, mainHeight, newLogLines int

	profileCPU *os.File

	buttonKeybindRotateCCW *tview.Button
	buttonKeybindRotateCW  *tview.Button
	buttonKeybindMoveLeft  *tview.Button
	buttonKeybindMoveRight *tview.Button
	buttonKeybindSoftDrop  *tview.Button
	buttonKeybindHardDrop  *tview.Button
	buttonKeybindCancel    *tview.Button
	buttonKeybindSave      *tview.Button

	buttonCancel *tview.Button
	buttonStart  *tview.Button
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
	renderLTee     = []byte(string(tcell.RuneLTee))
	renderRTee     = []byte(string(tcell.RuneRTee))
	renderULCorner = []byte(string(tcell.RuneULCorner))
	renderURCorner = []byte(string(tcell.RuneURCorner))
	renderLLCorner = []byte(string(tcell.RuneLLCorner))
	renderLRCorner = []byte(string(tcell.RuneLRCorner))
)

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

func handleResize(screen tcell.Screen) {
	newScreenW, newScreenH = screen.Size()
	if newScreenW == screenW && newScreenH == screenH {
		return
	}

	screenW, screenH = newScreenW, newScreenH

	if !fixedBlockSize {
		if screenW >= 80 && screenH >= 44 {
			blockSize = 2
		} else {
			blockSize = 1
		}
	}

	mainHeight = (20 * blockSize) + 2
	if screenH > mainHeight+9 {
		extraScreenPadding = 3
		mainHeight++
		inputHeight = 2
	} else if screenH > mainHeight+7 {
		extraScreenPadding = 2
		mainHeight++
		inputHeight = 2
	} else if screenH > mainHeight+5 {
		extraScreenPadding = 1
		mainHeight++
		inputHeight = 1
	} else if screenH > mainHeight+2 {
		extraScreenPadding = 0
		mainHeight++
		inputHeight = 1
	} else {
		extraScreenPadding = 0
		inputHeight = 1
	}

	multiplayerMatrixSize = ((screenW - extraScreenPadding) - ((10 * blockSize) + 16)) / ((10 * blockSize) + 4)

	newLogLines = ((screenH - mainHeight) - inputHeight) - extraScreenPadding
	if newLogLines > 0 {
		showLogLines = newLogLines
	} else {
		showLogLines = 1
	}

	gameGrid.SetRows(mainHeight+extraScreenPadding, inputHeight, -1).SetColumns(1+extraScreenPadding, 4+(10*blockSize), 10, -1)

	draw <- event.DrawAll
}

func drawAll() {
	if activeGame == nil {
		return
	}

	renderPlayerMatrix()
	renderPreviewMatrix()
	renderMultiplayerMatrix()
}

func drawMessages() {
	recent.ScrollToEnd()
}

func drawPlayerMatrix() {
	renderPlayerMatrix()
	renderPreviewMatrix()
}

func drawMultiplayerMatrixes() {
	renderMultiplayerMatrix()
}

func handleDraw() {
	var o event.DrawObject
	for o = range draw {
		switch o {
		case event.DrawMessages:
			app.QueueUpdateDraw(drawMessages)
		case event.DrawPlayerMatrix:
			app.QueueUpdateDraw(drawPlayerMatrix)
		case event.DrawMultiplayerMatrixes:
			app.QueueUpdateDraw(drawMultiplayerMatrixes)
		default:
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

	app.QueueUpdateDraw(func() {
		inputView.SetText("")
		if inputActive {
			app.SetFocus(inputView)
		} else {
			app.SetFocus(nil)
		}
	})
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
	if m.Combo > 0 && time.Until(m.ComboEnd) > 0 {
		comboTime = 1.0 + (float64(time.Until(m.ComboEnd)) / 1000000000)
		combo = m.Combo
	}

	var speed = strconv.Itoa(m.Speed)
	if m.Speed < 100 {
		speed = " " + speed
	}

	renderLock.Lock()
	renderMatrix(g.Players[g.LocalPlayer].Preview)

	if blockSize > 1 {
		renderBuffer.WriteString(fmt.Sprintf("\n\n\n\n\n Combo\n\n   %d\n\n\n\n\n Timer\n\n   %.0f\n\n\n\n\nPending\n\n   %d\n\n\n\n\n Speed\n\n  %s", combo, comboTime, m.PendingGarbage, speed))
	} else {
		renderBuffer.WriteString(fmt.Sprintf("\n\n Combo\n\n   %d\n\n Timer\n\n   %.0f\n\nPending\n\n   %d\n\n Speed\n\n  %s", combo, comboTime, m.PendingGarbage, speed))
	}

	side.Clear()
	side.Write(renderBuffer.Bytes())

	renderLock.Unlock()
	m.Unlock()
}

func renderPlayerMatrix() {
	g := activeGame
	if g == nil || len(g.Players) == 0 {
		return
	}

	renderLock.Lock()
	renderMatrix(g.Players[g.LocalPlayer].Matrix)
	mtx.Clear()
	mtx.Write(renderBuffer.Bytes())
	renderLock.Unlock()
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

	renderLock.Lock()
	renderMatrixes(matrixes)
	buffer.Clear()
	buffer.Write(renderBuffer.Bytes())
	renderLock.Unlock()
}

func renderMatrix(m *mino.Matrix) {
	renderBuffer.Reset()

	if m == nil {
		return
	}

	m.Lock()
	defer m.Unlock()

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

	for i := 0; i < extraScreenPadding; i++ {
		if m.Type == mino.MatrixStandard && i == extraScreenPadding-1 {
			renderBuffer.Write(renderULCorner)
			for x := 0; x < m.W*bs; x++ {
				renderBuffer.Write(renderHLine)
			}
			renderBuffer.Write(renderURCorner)
		}

		renderBuffer.WriteRune('\n')
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

			if y != 0 || m.Type != mino.MatrixCustom {
				renderBuffer.WriteRune('\n')
			}
		}
	}

	if m.Type != mino.MatrixStandard {
		return
	}

	renderBuffer.Write(renderLLCorner)
	for x := 0; x < m.W*bs; x++ {
		renderBuffer.Write(renderHLine)
	}
	renderBuffer.Write(renderLRCorner)

	renderBuffer.WriteRune('\n')
	renderPlayerDetails(m, bs)
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

func renderMatrixes(mx []*mino.Matrix) {
	renderBuffer.Reset()

	if mx == nil {
		return
	}

	for i := range mx {
		mx[i].Lock()
		mx[i].DrawPiecesL()
	}

	div := "  "

	height := mx[0].H

	for i := 0; i < extraScreenPadding; i++ {
		if i == extraScreenPadding-1 {
			for i := range mx {
				if i > 0 {
					renderBuffer.WriteString(div)
				}

				renderBuffer.Write(renderULCorner)
				for x := 0; x < mx[i].W*blockSize; x++ {
					renderBuffer.Write(renderHLine)
				}
				renderBuffer.Write(renderURCorner)
			}
		}

		renderBuffer.WriteRune('\n')
	}

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
}

func logMessage(message string) {
	logMutex.Lock()

	var prefix string
	if !wroteFirstLogMessage {
		wroteFirstLogMessage = true
	} else {
		prefix = "\n"
	}

	recent.Write([]byte(prefix + time.Now().Format(event.LogFormat) + " " + message))

	draw <- event.DrawMessages

	logMutex.Unlock()
}
