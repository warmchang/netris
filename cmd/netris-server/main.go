package main

import (
	"flag"
	"log"

	"git.sr.ht/~tslocum/netris/pkg/player"
	"git.sr.ht/~tslocum/netris/pkg/player/ssh"
)

func init() {
	log.SetFlags(0)
}

func main() {
	flag.Parse()

	s := &ssh.SSHServer{}

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
