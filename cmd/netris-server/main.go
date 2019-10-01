package main

import (
	"flag"
	"log"

	"git.sr.ht/~tslocum/netris/pkg/player"
	"git.sr.ht/~tslocum/netris/pkg/player/ssh"
)

var (
	listenAddressSSH string
	netrisPath       string
)

func init() {
	log.SetFlags(0)

	flag.StringVar(&listenAddressSSH, "listen-ssh", "", "SSH server listen address")
	flag.StringVar(&netrisPath, "netris", "", "path to netris")
}

func main() {
	flag.Parse()

	s := &ssh.SSHServer{ListenAddress: listenAddressSSH, NetrisPath: netrisPath}

	ps := player.NewServer([]player.ServerInterface{s})

	select {}
	_ = ps
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
