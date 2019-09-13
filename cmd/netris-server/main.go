package main

import (
	"flag"
	"log"
	"strconv"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
}

func main() {
	flag.Parse()

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
	}
}
