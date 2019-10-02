package game

import (
	"bufio"
	"log"
	"net"

	"git.sr.ht/~tslocum/netris/pkg/player"
)

type ServerConn struct {
	Conn net.Conn

	In  chan player.GameCommand
	Out chan player.GameCommand
}

func ConnectUnix(path string) *ServerConn {
	conn, err := net.Dial("unix", path)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	s := ServerConn{Conn: conn}

	in := make(chan player.GameCommand, player.CommandQueueSize)
	out := make(chan player.GameCommand, player.CommandQueueSize)

	s.In = in
	s.Out = out

	go s.handleRead()
	go s.handleWrite()

	return &s
}

func (s *ServerConn) handleRead() {
	scanner := bufio.NewScanner(s.Conn)
	for scanner.Scan() {
		line := scanner.Text()

		log.Println("read server conn ", line)
	}
}

func (s *ServerConn) handleWrite() {
	for e := range s.Out {
		_ = e
		log.Println("write", e)
		s.Conn.Write([]byte("test\n"))
	}
}
