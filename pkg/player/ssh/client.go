package ssh

import (
	"bufio"
	"io"
	"log"
	"strconv"

	"git.sr.ht/~tslocum/netris/pkg/player"
	"github.com/gliderlabs/ssh"
)

type SSHClient struct {
	S          ssh.Session
	Terminated chan bool

	In  chan<- player.GameCommand
	Out <-chan player.GameCommand
}

func NewSSHClient(s ssh.Session) *SSHClient {
	c := &SSHClient{S: s, Terminated: make(chan bool)}

	go c.handleIncoming()

	return c
}

func (c *SSHClient) handleIncoming() {
	r := bufio.NewScanner(c.S)
	for r.Scan() {
		c.In <- player.GameCommand{C: player.CommandChat, P: "sent " + r.Text()}
	}
	c.Detach("Disconnected")
}

func (c *SSHClient) handleWrite() {
	for w := range c.Out {
		io.WriteString(c.S, strconv.Itoa(int(w.C)))
	}
	c.Detach("Disconnected")
}

func (c *SSHClient) Attach(in chan<- player.GameCommand, out <-chan player.GameCommand) {
	c.In = in
	c.Out = out
}

func (c *SSHClient) Detach(reason string) {
	log.Println("DETACH", reason)
	c.Terminated <- true
}
