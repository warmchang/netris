package ssh

import (
	"log"
	"os"
	"path"

	"git.sr.ht/~tslocum/netris/pkg/player"
	"github.com/gliderlabs/ssh"
)

type SSHServer struct {
}

func (s *SSHServer) Host(newPlayers chan<- *player.ConnectingPlayer) {
	ssh.Handle(func(s ssh.Session) {
		c := NewSSHClient(s)

		newPlayers <- &player.ConnectingPlayer{Client: c, Name: s.User()}

		<-c.Terminated
	})

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		return true
	})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	go log.Fatal(ssh.ListenAndServe("localhost:7777", nil, ssh.HostKeyFile(path.Join(homeDir, ".ssh", "id_rsa")), publicKeyOption))
}

func (s *SSHServer) Shutdown(reason string) {
	// Stop listening
}
