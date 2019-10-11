package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/mino"
	"github.com/mattn/go-isatty"
)

var (
	ready = make(chan bool)
	done  = make(chan bool)

	activeGame *game.Game

	connectAddress string
	debugAddress   string
	nickname       string
	startMatrix    string

	blockSize = 1

	logDebug   bool
	logVerbose bool

	logMessages       []string
	renderLogMessages bool
	logMutex          = new(sync.Mutex)
	showLogLines      = 7
)

const (
	LogTimeFormat = "3:04:05"
)

func init() {
	log.SetFlags(0)
}
func main() {
	defer func() {
		if r := recover(); r != nil {
			closeGUI()

			time.Sleep(time.Second)

			log.Println()
			log.Println()
			log.Println()
			log.Println()
			debug.PrintStack()
			log.Fatalf("panic: %+v", r)
			os.Exit(0)
		}
	}()

	flag.IntVar(&blockSize, "size", 1, "block size")
	flag.StringVar(&nickname, "nick", "Anonymous", "nickname")
	flag.StringVar(&startMatrix, "matrix", "", "pre-fill matrix with pieces")
	flag.StringVar(&connectAddress, "connect", "", "server address to connect to")
	flag.StringVar(&debugAddress, "debug-address", "", "address to serve debug info")
	flag.BoolVar(&logDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&logVerbose, "verbose", false, "enable verbose logging")
	flag.Parse()

	// TODO Document
	if blockSize > 3 {
		blockSize = 3
	}

	logLevel := game.LogStandard
	if logVerbose {
		logLevel = game.LogVerbose
	} else if logDebug {
		logLevel = game.LogDebug
	}

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start netris: non-interactive terminals are not supported")
	}

	app, err := initGUI()
	if err != nil {
		log.Fatalf("failed to initialize GUI: %s", err)
	}

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("failed to run application: %s", err)
		}

		done <- true
	}()

	inputActive = true
	setInputStatus(false)

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			logMutex.Lock()
			logMessages = append(logMessages, time.Now().Format(LogTimeFormat)+" "+msg)
			renderLogMessages = true
			logMutex.Unlock()
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-sigc

		done <- true
	}()

	// Connect to a game
	if connectAddress != "" {
		s := game.ConnectUnix("/tmp/netris.sock")

		activeGame, err = s.JoinGame(nickname, 0, logger, draw)
		if err != nil {
			panic(err)
		}

		activeGame.LogLevel = logLevel

		<-done

		closeGUI()
		return
	}

	// Host a game
	server := game.NewServer(nil)

	server.Logger = make(chan string, game.LogQueueSize)
	go func() {
		for msg := range server.Logger {
			logMutex.Lock()
			logMessages = append(logMessages, time.Now().Format(LogTimeFormat)+" Local server: "+msg)
			renderLogMessages = true
			logMutex.Unlock()
		}
	}()

	go server.ListenUnix("/tmp/netris.sock")

	localServerConn := game.ConnectUnix("/tmp/netris.sock")

	activeGame, err = localServerConn.JoinGame(nickname, -1, logger, draw)
	if err != nil {
		panic(err)
	}

	activeGame.LogLevel = logLevel

	if startMatrix != "" {
		activeGame.Players[activeGame.LocalPlayer].Matrix.Lock()
		startMatrixSplit := strings.Split(startMatrix, ",")
		startMatrix = ""
		var (
			token int
			x     int
			err   error
		)
		for i := range startMatrixSplit {
			token, err = strconv.Atoi(startMatrixSplit[i])
			if err != nil {
				panic(fmt.Sprintf("failed to parse initial matrix on token #%d", i))
			}
			if i%2 == 1 {
				activeGame.Players[activeGame.LocalPlayer].Matrix.SetBlock(x, token, mino.BlockGarbage, false)
			} else {
				x = token
			}
		}
		activeGame.Players[activeGame.LocalPlayer].Matrix.Unlock()
	}

	<-done

	server.StopListening()

	closeGUI()
}
