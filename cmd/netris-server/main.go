package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.rocket9labs.com/tslocum/netris/pkg/event"
	"code.rocket9labs.com/tslocum/netris/pkg/game"
	"code.rocket9labs.com/tslocum/netris/pkg/game/ssh"
)

var (
	listenAddressTCP    string
	listenAddressSocket string
	listenAddressSSH    string
	netrisBinary        string
	debugAddress        string

	logDebug   bool
	logVerbose bool

	done = make(chan bool)
)

func init() {
	log.SetFlags(0)

	flag.StringVar(&listenAddressTCP, "listen-tcp", "", "host server on network address")
	flag.StringVar(&listenAddressSocket, "listen-socket", "", "host server on socket path")
	flag.StringVar(&listenAddressSSH, "listen-ssh", "", "host SSH server on network address")
	flag.StringVar(&netrisBinary, "netris", "", "path to netris client")
	flag.StringVar(&debugAddress, "debug-address", "", "address to serve debug info")
	flag.BoolVar(&logDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&logVerbose, "verbose", false, "enable verbose logging")
}

func main() {
	flag.Parse()

	if listenAddressTCP == "" && listenAddressSocket == "" {
		log.Fatal("at least one listen path or address is required (--listen-tcp and/or --listen-socket)")
	}

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	netrisAddress := listenAddressSocket
	if netrisAddress == "" {
		netrisAddress = listenAddressTCP
	}

	sshServer := &ssh.SSHServer{ListenAddress: listenAddressSSH, NetrisBinary: netrisBinary, NetrisAddress: netrisAddress}

	logLevel := game.LogStandard
	if logVerbose {
		logLevel = game.LogVerbose
	} else if logDebug {
		logLevel = game.LogDebug
	}

	server := game.NewServer([]game.ServerInterface{sshServer}, logLevel)

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			log.Println(time.Now().Format(event.LogFormat) + " " + msg)
		}
	}()

	server.Logger = logger

	if listenAddressSocket != "" {
		go server.Listen(listenAddressSocket)
	}
	if listenAddressTCP != "" {
		go server.Listen(listenAddressTCP)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-sigc

		done <- true
	}()

	<-done

	server.StopListening()
}
