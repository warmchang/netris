//+build !windows

package ssh

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"code.rocket9labs.com/tslocum/netris/pkg/game"
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

			configPath, err := createTemporaryConfig()
			if err != nil {
				log.Printf("warning: failed to create temporary configuration file: %s", err)
				return
			}

			cmdCtx, cancelCmd := context.WithCancel(sshSession.Context())
			defer cancelCmd()

			cmd := exec.CommandContext(cmdCtx, s.NetrisBinary, "--nick", game.Nickname(sshSession.User()), "--server", s.NetrisAddress, "--config", configPath)

			cmd.Env = append(sshSession.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))

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

			f.Close()
			cmd.Wait()
			os.Remove(configPath)
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

	hostKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	if _, err = os.Stat(hostKeyFile); os.IsNotExist(err) {
		log.Fatalf("failed to start SSH server: host key file %s not found\nto generate the missing key file, execute the following command:\nssh-keygen -t rsa -b 4096", hostKeyFile)
	}
	err = server.SetOption(ssh.HostKeyFile(hostKeyFile))
	if err != nil {
		log.Fatalf("failed to start SSH server: failed to set host key file %s: %s", hostKeyFile, err)
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

// Eventually this will load and save configuration based on public key hash.
func createTemporaryConfig() (string, error) {
	f, err := ioutil.TempFile("", "netris-config-*.yaml")
	if err != nil {
		return "", err
	}

	f.Close()
	return filepath.Clean(f.Name()), err
}
