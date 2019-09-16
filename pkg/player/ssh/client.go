package ssh

import "git.sr.ht/~tslocum/netris/pkg/player"

type SSHClient struct {
}

func (c *SSHClient) Attach(in <-chan player.GameCommand, out chan<- player.GameCommand) {

}

func (c *SSHClient) Detach(reason string) {
}
