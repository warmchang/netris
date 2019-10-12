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

	app        *tview.Application
	grid       *tview.Grid
	inputView  *tview.InputField
	mtx        *tview.TextView
	side       *tview.TextView
	buffer     *tview.TextView
	recent     *tview.TextView
	lowerPages *tview.Pages

	draw = make(chan event.DrawObject, game.CommandQueueSize)

	renderLock   = new(sync.Mutex)
	renderBuffer bytes.Buffer

	screenW, screenH       int
	newScreenW, newScreenH int
)

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

	if blockSize == 0 {
		app.SetBeforeDrawFunc(handleResize)
	}

	inputView = tview.NewInputField().
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

	grid = tview.NewGrid().
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

	recent.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		logMutex.Lock()
		showLogLines = height
		if showLogLines < 1 {
			showLogLines = 1
		}
		logMutex.Unlock()

		return recent.GetInnerRect()
	})

	lowerPages = tview.NewPages()
	lowerPages = lowerPages.AddPage("input",
		inputView,
		true, false)
	lowerPages = lowerPages.AddPage("recent",
		recent,
		true, true)

	grid = grid.SetColumns(1, 4+(10*blockSize), 10, -1).
		AddItem(spacer, 0, 0, 2, 1, 0, 0, false).
		AddItem(mtx, 0, 1, 1, 1, 0, 0, false).
		AddItem(side, 0, 2, 1, 1, 0, 0, false).
		AddItem(buffer, 0, 3, 1, 1, 0, 0, false).
		AddItem(lowerPages, 1, 1, 1, 3, 0, 0, true)

	app = app.SetInputCapture(handleKeypress)

	app = app.SetRoot(grid, true)

	go handleDraw()

	return app, nil
}

func handleResize(screen tcell.Screen) bool {
	newScreenW, newScreenH = screen.Size()
	if newScreenW != screenW || newScreenH != screenH {
		screenW, screenH = newScreenW, newScreenH

		oldBlockSize := blockSize
		if screenW >= 80 && screenH >= 44 {
			blockSize = 2
		} else {
			blockSize = 1
		}
		if blockSize == oldBlockSize {
			return false
		}

		grid.SetRows(2+(20*blockSize), -1).SetColumns(1, 4+(10*blockSize), 10, -1)

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

	if active {
		inputView.SetText("")
		lowerPages = lowerPages.SwitchToPage("input")
	} else {
		lowerPages = lowerPages.SwitchToPage("recent")
	}
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

	var matrixes []*mino.Matrix
	for _, playerID := range playerIDs {
		if g.Players[playerID] == nil {
			continue
		}

		matrixes = append(matrixes, g.Players[playerID].Matrix)
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

	// Draw preview matrix at block size 2 max
	bs := blockSize
	if m.Preview {
		if bs > 2 {
			bs = 2
		}
		if bs > 1 {
			renderBuffer.WriteRune('\n')
		}
	}

	for y := m.H - 1; y >= 0; y-- {
		for j := 0; j < bs; j++ {
			if !m.Preview {
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

			if !m.Preview {
				renderBuffer.Write(renderVLine)
			}

			renderBuffer.WriteRune('\n')
		}
	}

	if m.Preview {
		return renderBuffer.Bytes()
	}

	renderBuffer.Write(renderLLCorner)
	for x := 0; x < m.W*bs; x++ {
		renderBuffer.Write(renderHLine)
	}
	renderBuffer.Write(renderLRCorner)

	renderBuffer.WriteRune('\n')
	name := m.PlayerName
	if len(name) > m.W*bs {
		name = name[:m.W*bs]
	}
	padName := ((m.W*bs - len(name)) / 2) + 1
	for i := 0; i < padName; i++ {
		renderBuffer.WriteRune(' ')
	}
	renderBuffer.WriteString(name)

	return renderBuffer.Bytes()
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

				if !m.Preview {
					renderBuffer.Write(renderVLine)
				}

				for x := 0; x < m.W; x++ {
					for j := 0; j < blockSize; j++ {
						renderBuffer.Write(renderBlock[m.Block(x, y)])
					}
				}

				if !m.Preview {
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

		name := m.PlayerName
		if len(name) > m.W*blockSize {
			name = name[:m.W*blockSize]
		}
		padName := ((m.W*blockSize - len(name)) / 2) + 1
		for i := 0; i < padName; i++ {
			renderBuffer.WriteRune(' ')
		}
		renderBuffer.WriteString(name)
		padName = (m.W * blockSize) + 2 - len(name) - padName
		for i := 0; i < padName; i++ {
			renderBuffer.WriteRune(' ')
		}
	}

	for i := range mx {
		mx[i].Unlock()
	}

	return renderBuffer.Bytes()
}

func logMessage(message string) {
	logMutex.Lock()
	logMessages = append(logMessages, message)
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
