package main

import (
	"flag"
	"log"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func init() {
	log.SetFlags(0)
}

func main() {
	flag.Parse()

	minos, err := mino.Generate(5)
	if err != nil {
		panic(err)
	}
	for _, m := range minos {
		log.Println(m.Render())
	}
}
