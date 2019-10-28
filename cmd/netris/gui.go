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
	screenPadding         int

	screenW, screenH       int
	newScreenW, newScreenH int

	nickname      = "Anonymous"
	nicknameDraft string

	inputHeight, mainHeight, previewWidth, newLogLines int

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

// BS 1: 10x10
// BS 2: 20x20
// BS 3: 30x40
func handleResize(screen tcell.Screen) {
	newScreenW, newScreenH = screen.Size()
	if newScreenW == screenW && newScreenH == screenH {
		return
	}

	screenW, screenH = newScreenW, newScreenH

	if !fixedBlockSize {
		if screenW >= 80 && screenH >= 46 {
			blockSize = 3
		} else if screenW >= 80 && screenH >= 24 {
			blockSize = 2
		} else {
			blockSize = 1
		}
	}

	xMultiplier := 1
	if blockSize == 2 {
		xMultiplier = 2
	} else if blockSize == 3 {
		xMultiplier = 3
	}

	if blockSize == 1 {
		mainHeight = 10 + 3
	} else if blockSize == 2 {
		mainHeight = 20 + 3
	} else {
		mainHeight = 40 + 3
	}

	if screenH > mainHeight+9 {
		screenPadding = 2
		mainHeight++
		inputHeight = 2
	} else if screenH > mainHeight+7 {
		screenPadding = 2
		mainHeight++
		inputHeight = 1
	} else if screenH > mainHeight+5 {
		screenPadding = 1
		mainHeight++
		inputHeight = 1
	} else if screenH > mainHeight+3 {
		screenPadding = 1
		inputHeight = 1
	} else {
		screenPadding = 0
		inputHeight = 0
	}

	if blockSize == 1 {
		previewWidth = 9
	} else if blockSize == 2 {
		previewWidth = 10
	} else {
		previewWidth = 14
	}

	multiplayerMatrixSize = ((screenW - screenPadding) - ((10 * xMultiplier) + previewWidth + 6)) / ((10 * xMultiplier) + 4)

	newLogLines = ((screenH - mainHeight) - inputHeight) - screenPadding
	if newLogLines > 0 {
		showLogLines = newLogLines
	} else {
		showLogLines = 1
	}

	gameGrid.SetRows(screenPadding, mainHeight, inputHeight, -1).SetColumns(screenPadding+1, 4+(10*xMultiplier), previewWidth, -1)

	draw <- event.DrawAll
}

func drawAll() {
	if activeGame == nil {
		return
	}

	renderPlayerGUI()
	renderMultiplayerGUI()
}

func drawMessages() {
	recent.ScrollToEnd()
}

func drawPlayerMatrix() {
	renderPlayerGUI()
}

func drawMultiplayerMatrixes() {
	renderMultiplayerGUI()
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

func renderPlayerGUI() {
	g := activeGame
	if g == nil || len(g.Players) == 0 {
		return
	}

	renderLock.Lock()
	renderMatrixes([]*mino.Matrix{g.Players[g.LocalPlayer].Matrix})
	mtx.Clear()
	mtx.Write(renderBuffer.Bytes())
	renderLock.Unlock()

	player := g.Players[g.LocalPlayer]
	m := g.Players[g.LocalPlayer].Matrix

	p := mino.NewPiece(m.Bag.Next(), mino.Point{0, 0})

	player.Preview.Clear()

	if !player.Matrix.GameOver {
		err := player.Preview.Add(p, p.Solid, mino.Point{0, 0}, false)
		if err != nil {
			log.Fatalf("failed to render preview matrix: failed to add ghost piece: %+v", err)
		}
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
	renderMatrixes([]*mino.Matrix{g.Players[g.LocalPlayer].Preview})

	if blockSize == 1 {
		renderBuffer.WriteString(fmt.Sprintf(" Combo\n   %d\n\n Timer\n   %.0f\n\nPending\n   %d\n\n Speed\n  %s", combo, comboTime, m.PendingGarbage, speed))
	} else if blockSize == 2 {
		renderBuffer.WriteString(fmt.Sprintf("\n Combo\n\n   %d\n\n\n Timer\n\n   %.0f\n\n\nPending\n\n   %d\n\n\n Speed\n\n  %s", combo, comboTime, m.PendingGarbage, speed))
	} else if blockSize == 3 {
		renderBuffer.WriteString(fmt.Sprintf("\n\n\n\n\n   Combo\n\n     %d\n\n\n\n\n\n   Timer\n\n     %.0f\n\n\n\n\n\n  Pending\n\n     %d\n\n\n\n\n\n   Speed\n\n    %s", combo, comboTime, m.PendingGarbage, speed))
	}

	side.Clear()
	side.Write(renderBuffer.Bytes())

	renderLock.Unlock()
	m.Unlock()
}

func renderMultiplayerGUI() {
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
	if mx == nil || len(mx) == 0 {
		return
	}

	bs := blockSize
	mt := mx[0].Type
	mh := mx[0].H
	div := "  "

	var nextPieceWidth = 0
	if mt == mino.MatrixPreview {
		renderBuffer.WriteRune('\n')
		if mx[0].Bag != nil {
			p := mx[0].Bag.Next()
			nextPieceWidth, _ = p.Size()
			if nextPieceWidth == 2 && blockSize == 1 {
				nextPieceWidth = 3
			}
		}
	} else if mt == mino.MatrixCustom {
		bs = 1
	}

	for i := range mx {
		mx[i].Lock()
		mx[i].DrawPiecesL()
	}

	if mt == mino.MatrixStandard {
		for i := range mx {
			if i > 0 {
				renderBuffer.WriteString(div)
			}

			renderBuffer.Write(renderULCorner)
			for x := 0; x < mx[i].W*bs; x++ {
				renderBuffer.Write(renderHLine)
			}
			renderBuffer.Write(renderURCorner)
		}
		renderBuffer.WriteRune('\n')
	}

	if bs == 1 {
		for y := mh - 1; y >= 0; y -= 2 {
			for i, m := range mx {
				if i > 0 {
					renderBuffer.WriteString(div)
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				} else if m.Type == mino.MatrixPreview {
					renderBuffer.WriteRune(' ')

					if nextPieceWidth < 4 {
						renderBuffer.WriteRune(' ')
					}
				}

				for x := 0; x < m.W; x++ {
					renderBuffer.WriteRune('[')
					renderBuffer.Write(mino.Colors[m.Block(x, y-1)])
					renderBuffer.WriteRune(':')
					renderBuffer.Write(mino.Colors[m.Block(x, y)])
					renderBuffer.WriteRune(']')
					renderBuffer.WriteRune('▄')
					renderBuffer.Write([]byte("[-:-]"))
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				}
			}

			if y > 1 || mt != mino.MatrixCustom {
				renderBuffer.WriteRune('\n')
			}
		}
	} else if bs == 2 {
		for y := mh - 1; y >= 0; y-- {
			for i, m := range mx {
				if i > 0 {
					renderBuffer.WriteString(div)
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				} else if m.Type == mino.MatrixPreview {
					for pad := 0; pad < 3-nextPieceWidth; pad++ {
						renderBuffer.WriteRune(' ')
						renderBuffer.WriteRune(' ')
					}
				}

				for x := 0; x < m.W; x++ {
					renderBuffer.WriteRune('[')
					renderBuffer.Write(mino.Colors[m.Block(x, y)])
					renderBuffer.WriteRune(']')
					renderBuffer.WriteRune('█')
					renderBuffer.WriteRune('█')
					renderBuffer.Write([]byte("[-]"))
				}

				if m.Type == mino.MatrixStandard {
					renderBuffer.Write(renderVLine)
				}
			}

			if y != 0 || mt != mino.MatrixCustom {
				renderBuffer.WriteRune('\n')
			}
		}
	} else {
		for y := mh - 1; y >= 0; y-- {
			for repeat := 0; repeat < 2; repeat++ {
				for i, m := range mx {
					if i > 0 {
						renderBuffer.WriteString(div)
					}

					if m.Type == mino.MatrixStandard {
						renderBuffer.Write(renderVLine)
					} else if m.Type == mino.MatrixPreview {
						if nextPieceWidth == 2 {
							renderBuffer.WriteRune(' ')
							renderBuffer.WriteRune(' ')
						} else if nextPieceWidth < 4 {
							renderBuffer.WriteRune(' ')
						}
					}

					for x := 0; x < m.W; x++ {
						renderBuffer.WriteRune('[')
						renderBuffer.Write(mino.Colors[m.Block(x, y)])
						renderBuffer.WriteRune(']')
						renderBuffer.WriteRune('█')
						renderBuffer.WriteRune('█')
						renderBuffer.WriteRune('█')
						renderBuffer.Write([]byte("[-]"))
					}

					if m.Type == mino.MatrixStandard {
						renderBuffer.Write(renderVLine)
					}
				}

				if y != 0 || mt != mino.MatrixCustom {
					renderBuffer.WriteRune('\n')
				}
			}
		}
	}

	if mt == mino.MatrixStandard {
		for i := range mx {
			if i > 0 {
				renderBuffer.WriteString(div)
			}

			renderBuffer.Write(renderLLCorner)
			for x := 0; x < mx[i].W*bs; x++ {
				renderBuffer.Write(renderHLine)
			}
			renderBuffer.Write(renderLRCorner)
		}

		renderBuffer.WriteRune('\n')

		for i, m := range mx {
			if i > 0 {
				renderBuffer.WriteString(div)
			}

			renderPlayerDetails(m, bs)
		}
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
