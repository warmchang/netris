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

	"gitlab.com/tslocum/cview"
	"gitlab.com/tslocum/netris/pkg/event"
	"gitlab.com/tslocum/netris/pkg/game"
	"gitlab.com/tslocum/netris/pkg/mino"
)

var (
	closedGUI bool

	inputActive      bool
	capturingKeybind bool
	showDetails      bool

	app       *cview.Application
	inputView *cview.InputField
	mtx       *cview.TextView
	side      *cview.TextView
	buffer    *cview.TextView
	recent    *cview.TextView

	joinedGame bool

	draw     = make(chan event.DrawObject, game.CommandQueueSize)
	joinGame = make(chan int, game.CommandQueueSize)

	renderLock   = new(sync.Mutex)
	renderBuffer bytes.Buffer

	multiplayerMatrixSize int
	screenPadding         int

	screenW, screenH int

	nickname = "Anonymous"

	drawGhostPiece        = true
	drawGhostPieceUnsaved bool

	inputHeight, mainHeight, previewWidth, newLogLines int

	profileCPU *os.File

	playerSettingsCancel *cview.Button
	playerSettingsSave   *cview.Button

	buttonGhostPiece       *cview.Button
	buttonKeybindRotateCCW *cview.Button
	buttonKeybindRotateCW  *cview.Button
	buttonKeybindMoveLeft  *cview.Button
	buttonKeybindMoveRight *cview.Button
	buttonKeybindSoftDrop  *cview.Button
	buttonKeybindHardDrop  *cview.Button
	buttonKeybindCancel    *cview.Button
	buttonKeybindSave      *cview.Button

	buttonNewGameCancel *cview.Button
	buttonNewGameStart  *cview.Button
)

const DefaultStatusText = "Press Enter to chat, Z/X to rotate, arrow keys or HJKL to move/drop"

var (
	renderHLine    []byte
	renderVLine    []byte
	renderLTee     []byte
	renderRTee     []byte
	renderULCorner []byte
	renderURCorner []byte
	renderLLCorner []byte
	renderLRCorner []byte
)

func setBorderColor(color string) {
	singleChar := []byte(fmt.Sprintf("[-:%s] [-:-]", color))
	doubleChar := []byte(fmt.Sprintf("[-:%s]  [-:-]", color))

	renderHLine = singleChar
	renderVLine = doubleChar
	renderLTee = singleChar
	renderRTee = singleChar
	renderULCorner = doubleChar
	renderURCorner = doubleChar
	renderLLCorner = doubleChar
	renderLRCorner = doubleChar
}

func resetPlayerSettingsForm() {
	playerSettingsNameInput.SetText(nickname)
}

// BS 1: 10x10
// BS 2: 20x20
// BS 3: 40x40
func handleResize(width int, height int) {
	if width == screenW && height == screenH {
		return
	}

	screenW, screenH = width, height

	if !fixedBlockSize {
		if screenW >= 106 && screenH >= 46 {
			blockSize = 3
		} else if screenW >= 56 && screenH >= 24 {
			blockSize = 2
		} else {
			blockSize = 1
		}
	}

	xMultiplier := 1
	if blockSize == 2 {
		xMultiplier = 2
	} else if blockSize == 3 {
		xMultiplier = 4
	}

	if blockSize == 1 {
		mainHeight = 10 + 3
	} else if blockSize == 2 {
		mainHeight = 20 + 3
	} else {
		mainHeight = 40 + 3
	}

	if screenH > mainHeight+9 {
		screenPadding = 1
		mainHeight++
		inputHeight = 2
	} else if screenH > mainHeight+7 {
		screenPadding = 1
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
		previewWidth = 18
	}

	multiplayerMatrixSize = ((screenW - screenPadding) - ((10 * xMultiplier) + previewWidth + 6)) / ((10 * xMultiplier) + 6)

	newLogLines = ((screenH - mainHeight) - inputHeight) - screenPadding
	if newLogLines > 0 {
		showLogLines = newLogLines
	} else {
		showLogLines = 1
	}

	gameGrid.SetRows(screenPadding, mainHeight, inputHeight, -1).SetColumns(screenPadding+1, 5+(10*xMultiplier), previewWidth, -1)

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
			log.Fatalf("failed to render preview matrix: failed to add preview piece: %+v", err)
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
		renderBuffer.WriteString(fmt.Sprintf("\n\n\n\n\n    Combo\n\n      %d\n\n\n\n\n\n    Timer\n\n      %.0f\n\n\n\n\n\n   Pending\n\n      %d\n\n\n\n\n\n    Speed\n\n     %s", combo, comboTime, m.PendingGarbage, speed))
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
	xMultiplier := 1
	if bs == 2 {
		xMultiplier = 2
	} else if bs == 3 {
		xMultiplier = 4
	}

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
	if len(buf) > m.W*xMultiplier {
		buf = buf[:m.W*xMultiplier]
	}

	padBuf := ((m.W*xMultiplier - len(buf)) / 2) + 3
	for i := 0; i < padBuf; i++ {
		renderBuffer.WriteRune(' ')
	}
	renderBuffer.WriteString(buf)
	padBuf = m.W*xMultiplier + 4 - len(buf) - padBuf
	for i := 0; i < padBuf; i++ {
		renderBuffer.WriteRune(' ')
	}
}

func renderMatrixes(mx []*mino.Matrix) {
	renderBuffer.Reset()
	if len(mx) == 0 {
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
			if p != nil {
				nextPieceWidth, _ = p.Size()
			}
		}

		if bs == 3 {
			renderBuffer.WriteRune('\n')
		}
	} else if mt == mino.MatrixCustom {
		bs = 1
	}

	xMultiplier := 1
	if bs == 2 {
		xMultiplier = 2
	} else if bs == 3 {
		xMultiplier = 4
	}

	for i := range mx {
		mx[i].Lock() // Unlocked later in this function

		if mt == mino.MatrixCustom {
			continue
		}

		mx[i].ClearOverlayL()
		if drawGhostPiece {
			mx[i].DrawGhostPieceL()
		}
		mx[i].DrawActivePieceL()
	}

	if mt == mino.MatrixStandard {
		for i := range mx {
			if i > 0 {
				renderBuffer.WriteString(div)
			}

			renderBuffer.Write(renderULCorner)
			for x := 0; x < mx[i].W*xMultiplier; x++ {
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
				}

				for x := 0; x < m.W; x++ {
					if m.Block(x, y-1) == mino.BlockNone && m.Block(x, y) == mino.BlockNone {
						renderBuffer.WriteRune(' ')
						continue
					} else if m.Block(x, y-1) == mino.BlockNone {
						renderBuffer.WriteRune('[')
						renderBuffer.Write(mino.Colors[m.Block(x, y)])
						renderBuffer.WriteRune(']')
						renderBuffer.WriteRune('▀')
						renderBuffer.Write([]byte("[-:-]"))
						continue
					} else if m.Block(x, y) == mino.BlockNone {
						renderBuffer.WriteRune('[')
						renderBuffer.Write(mino.Colors[m.Block(x, y-1)])
						renderBuffer.WriteRune(']')
						renderBuffer.WriteRune('▄')
						renderBuffer.Write([]byte("[-:-]"))
						continue
					}

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
					if nextPieceWidth < 4 {
						renderBuffer.WriteRune(' ')
					}
				}

				for x := 0; x < m.W; x++ {
					if m.Block(x, y) == mino.BlockNone {
						renderBuffer.WriteRune(' ')
						renderBuffer.WriteRune(' ')
						continue
					}

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
						if nextPieceWidth < 4 {
							renderBuffer.WriteRune(' ')
						}
					}

					for x := 0; x < m.W; x++ {
						if m.Block(x, y) == mino.BlockNone {
							renderBuffer.WriteRune(' ')
							renderBuffer.WriteRune(' ')
							renderBuffer.WriteRune(' ')
							renderBuffer.WriteRune(' ')
							continue
						}

						renderBuffer.WriteRune('[')
						renderBuffer.Write(mino.Colors[m.Block(x, y)])
						renderBuffer.WriteRune(']')
						renderBuffer.WriteRune('█')
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
			for x := 0; x < mx[i].W*xMultiplier; x++ {
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
