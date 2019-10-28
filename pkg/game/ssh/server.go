//+build !windows

package ssh

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
	"unsafe"

	"git.sr.ht/~tslocum/netris/pkg/game"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

const (
	ServerIdleTimeout = 1 * time.Minute
)

var server *ssh.Server

type SSHServer struct {
	ListenAddress string
	NetrisBinary  string
	NetrisAddress string
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (s *SSHServer) Host(newPlayers chan<- *game.IncomingPlayer) {
	if s.ListenAddress == "" {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to retrieve user home dir: %s", err)
	}

	server = &ssh.Server{
		Addr:        s.ListenAddress,
		IdleTimeout: ServerIdleTimeout,
		Handler: func(sshSession ssh.Session) {
			ptyReq, winCh, isPty := sshSession.Pty()
			if !isPty {
				io.WriteString(sshSession, "failed to start netris: non-interactive terminals are not supported\n")

				sshSession.Exit(1)
				return
			}

			cmdCtx, cancelCmd := context.WithCancel(sshSession.Context())

			cmd := exec.CommandContext(cmdCtx, s.NetrisBinary, "--nick", game.Nickname(sshSession.User()), "--server", s.NetrisAddress)
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))

			f, err := pty.Start(cmd)
			if err != nil {
				io.WriteString(sshSession, fmt.Sprintf("failed to start netris: failed to initialize pseudo-terminal: %s\n", err))

				sshSession.Exit(1)
				return
			}
			defer f.Close()

			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()

			go func() {
				io.Copy(f, sshSession)
			}()
			io.Copy(sshSession, f)

			cancelCmd()
			cmd.Wait()
		},
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			// TODO: Compare public key

			return true
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			return true
		},
		KeyboardInteractiveHandler: func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
			return true
		},
	}

	err = server.SetOption(ssh.HostKeyFile(path.Join(homeDir, ".ssh", "id_rsa")))
	if err != nil {
		log.Fatalf("failed to start SSH server: failed to set host key file: %s", err)
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("failed to start SSH server: %s", err)
		}
	}()
}

func (s *SSHServer) Shutdown(reason string) {
	server.Close()
}
