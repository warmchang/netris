package main

import (
	"flag"
	"log"
	"strconv"

	"git.sr.ht/~tslocum/netris/pkg/matrix"
	"git.sr.ht/~tslocum/netris/pkg/mino"
)

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
}

func main() {
	flag.Parse()

	m := matrix.NewMatrix(10, 20, 20)
	m.M[m.I(0, 2)] = mino.Solid
	m.M[m.I(4, 5)] = mino.Garbage
	//m.Print()

	i, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	minos := mino.Generate(i)
	for _, m := range minos {
		log.Println(m.Render())
		log.Println()
		log.Println()
	}
}
