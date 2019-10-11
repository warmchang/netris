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

	"git.sr.ht/~tslocum/netris/pkg/game"
	"git.sr.ht/~tslocum/netris/pkg/game/ssh"
)

var (
	listenAddressSSH string
	netrisPath       string
	debugAddress     string
	done             = make(chan bool)
)

const (
	LogTimeFormat = "2006-01-02 15:04:05"
)

func init() {
	log.SetFlags(0)

	flag.StringVar(&listenAddressSSH, "listen-ssh", "", "SSH server listen address")
	flag.StringVar(&netrisPath, "netris", "", "path to netris")
	flag.StringVar(&debugAddress, "debug", "", "address to serve debug info")
}

func main() {
	flag.Parse()

	if debugAddress != "" {
		go func() {
			log.Fatal(http.ListenAndServe(debugAddress, nil))
		}()
	}

	sshServer := &ssh.SSHServer{ListenAddress: listenAddressSSH, NetrisPath: netrisPath}

	server := game.NewServer([]game.ServerInterface{sshServer})

	logger := make(chan string, game.LogQueueSize)
	go func() {
		for msg := range logger {
			log.Println(time.Now().Format(LogTimeFormat) + " " + msg)
		}
	}()

	server.Logger = logger

	go server.ListenUnix("/tmp/netris.sock")

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

	/*
		i, err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			panic(err)
		}

		minos, err := mino.Generate(i)
		if err != nil {
			panic(err)
		}
		for _, m := range minos {
			log.Println(m.Render())
			log.Println()
			log.Println()
		}*/
}
