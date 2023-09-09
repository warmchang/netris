package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.rocket9labs.com/tslocum/netris/pkg/event"
	"code.rocket9labs.com/tslocum/netris/pkg/game"
	"code.rocket9labs.com/tslocum/netris/pkg/mino"
	"code.rocketnine.space/tslocum/ez"
	"github.com/mattn/go-isatty"
)

var (
	done = make(chan bool)

	activeGame     *game.Game
	activeGameConn *game.Conn

	server         *game.Server
	localListenDir string

	connectAddress string
	serverAddress  string
	debugAddress   string
	startMatrix    string

	nicknameFlag string

	configPath string

	blockSize      = 0
	fixedBlockSize bool

	logDebug   bool
	logVerbose bool

	logMutex             = new(sync.Mutex)
	wroteFirstLogMessage bool
	showLogLines         = 7
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
			debug.PrintStack()

			log.Println()
			log.Println()
			log.Fatalf("caught panic: %+v", r)
		}
	}()

	flag.IntVar(&blockSize, "scale", 0, "UI scale")
	flag.StringVar(&nicknameFlag, "nick", "", "nickname")
	flag.StringVar(&startMatrix, "matrix", "", "pre-fill matrix with pieces")
	flag.StringVar(&connectAddress, "connect", "", "connect to server address or socket path")
	flag.StringVar(&serverAddress, "server", game.DefaultServer, "server address or socket path")
	flag.StringVar(&debugAddress, "debug-address", "", "address to serve debug info")
	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.BoolVar(&logDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&logVerbose, "verbose", false, "enable verbose logging")
	flag.Parse()

	tty := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !tty {
		log.Fatal("failed to start netris: non-interactive terminals are not supported")
	}

	if blockSize > 0 {
		fixedBlockSize = true

		if blockSize > 3 {
			blockSize = 3
		}
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

	if configPath == "" {
		var err error
		configPath, err = ez.DefaultConfigPath("netris")
		if err != nil {
			log.Fatalf("failed to determine default configuration path: %s", err)
		}
	}

	err := ez.Deserialize(config, configPath)
	if err != nil {
		log.Fatalf("failed to read configuration file: %s", err)
	}

	err = setKeyBinds()
	if err != nil {
		log.Fatalf("failed to set keybinds: %s", err)
	}

	for gameColor, defaultColor := range event.DefaultColors {
		currentValue := strings.ToLower(config.Colors[gameColor])
		if currentValue == "" {
			currentValue = defaultColor
		} else if !regexpColor.MatchString(currentValue) {
			log.Fatalf("failed to set colors: invalid color provided for piece %s: %s", gameColor, currentValue)
		}
		config.Colors[gameColor] = currentValue

		blockColor := mino.ColorToBlock[gameColor]
		if blockColor > 0 {
			mino.Colors[blockColor] = []byte(currentValue)
		}
	}
	setBorderColor(config.Colors[event.GameColorBorder])

	if nicknameFlag != "" && game.Nickname(nicknameFlag) != "" {
		config.Name = game.Nickname(nicknameFlag)
	} else if config.Name != "" && game.Nickname(config.Name) != "" {
		config.Name = game.Nickname(config.Name)
	}

	app, err := initGUI(connectAddress != "")
	if err != nil {
		log.Fatalf("failed to initialize GUI: %s", err)
	}

	go func() {
		if err := app.Run(); err != nil {
			log.Fatalf("failed to run application: %s", err)
		}

		done <- true
	}()

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			logMessage(msg)
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

	// TODO Connect automatically when an address or path is supplied
	if connectAddress != "" {
		serverAddress = connectAddress
	} else {
		connectAddress = serverAddress
	}

	go func() {
		<-done

		if activeGameConn != nil {
			if server == nil {
				activeGameConn.Write(&game.GameCommandDisconnect{})
				activeGameConn.Wait()
			}

			activeGameConn.Close()
		}

		if server != nil {
			server.StopListening()
		}
		if localListenDir != "" {
			os.RemoveAll(localListenDir)
		}

		closeGUI()

		err := ez.Serialize(config, configPath)
		if err != nil {
			log.Printf("warning: failed to save configuration: %s", err)
		}

		os.Exit(0)
	}()

	for {
		gameID := <-joinGame

		if server != nil {
			server.StopListening()
			server = nil
		}
		if localListenDir != "" {
			os.RemoveAll(localListenDir)
			localListenDir = ""
		}

		if gameID == event.GameIDNewCustom || gameID >= 0 {
			joinedGame = true
			setTitleVisible(false)

			connectNetwork, _ := game.NetworkAndAddress(connectAddress)

			if connectNetwork != "unix" {
				logMessage(fmt.Sprintf("* Connecting to %s...", connectAddress))
			}

			activeGameConn, err = game.Connect(connectAddress)
			if err != nil {
				log.Fatal(err)
			}

			var newGame *game.ListedGame
			if gameID == event.GameIDNewCustom {
				gameID = 0

				maxPlayers, err := strconv.Atoi(newGameMaxPlayersInput.GetText())
				if err != nil {
					maxPlayers = 0
				}

				speedLimit, err := strconv.Atoi(newGameSpeedLimitInput.GetText())
				if err != nil {
					speedLimit = 0
				}

				newGame = &game.ListedGame{Name: game.GameName(newGameNameInput.GetText()), MaxPlayers: maxPlayers, SpeedLimit: speedLimit}
			}

			activeGame, err = activeGameConn.JoinGame(config.Name, gameID, newGame, logger, draw)
			if err != nil {
				log.Fatalf("failed to connect to %s: %s", connectAddress, err)
			}

			if activeGame == nil {
				log.Fatal("failed to connect to server")
			}

			activeGame.LogLevel = logLevel
			continue
		}

		joinedGame = true
		setTitleVisible(false)

		server = game.NewServer(nil, logLevel)

		server.Logger = make(chan string, game.LogQueueSize)
		if logDebug || logVerbose {
			go func() {
				for msg := range server.Logger {
					logMessage("Local server: " + msg)
				}
			}()
		} else {
			go func() {
				for range server.Logger {
				}
			}()
		}

		localListenDir, err = ioutil.TempDir("", "netris")
		if err != nil {
			log.Fatal(err)
		}

		localListenAddress := path.Join(localListenDir, "netris.sock")

		go server.Listen(localListenAddress)

		activeGameConn, err = game.Connect(localListenAddress)
		if err != nil {
			log.Fatalf("failed to create local game: %s", err)
		}

		activeGame, err = activeGameConn.JoinGame(config.Name, event.GameIDNewLocal, nil, logger, draw)
		if err != nil {
			log.Fatalf("failed to join local game: %s", err)
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
					log.Fatalf("failed to parse custom matrix on token #%d", i)
				}
				if i%2 == 1 {
					activeGame.Players[activeGame.LocalPlayer].Matrix.SetBlock(x, token, mino.BlockGarbage, false)
				} else {
					x = token
				}
			}
			activeGame.Players[activeGame.LocalPlayer].Matrix.Unlock()
		}
	}
}
