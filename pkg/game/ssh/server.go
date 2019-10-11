package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"
	"unsafe"

	"git.sr.ht/~tslocum/netris/pkg/game"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

const (
	ServerIdleTimeout = 1 * time.Minute
)

type SSHServer struct {
	ListenAddress string
	NetrisPath    string
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (s *SSHServer) Host(newPlayers chan<- *game.IncomingPlayer) {
	if s.ListenAddress == "" {
		panic("SSH server ListenAddress must be specified")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	server := &ssh.Server{
		Addr:        s.ListenAddress,
		IdleTimeout: ServerIdleTimeout,
		Handler: func(sshSession ssh.Session) {
			ctx := sshSession.Context()

			if publicKey, ok := ctx.Value("publickey").(ssh.PublicKey); ok {
				//log.Printf("logged in with %s", publicKey.Marshal())
				_ = publicKey
			}

			cmdCtx, cancelCmd := context.WithCancel(ctx)
			cmd := exec.CommandContext(cmdCtx, s.NetrisPath, "--nick", "'"+game.Nickname(sshSession.User())+"'", "--connect", "/tmp/netris.sock")
			ptyReq, winCh, isPty := sshSession.Pty()
			if isPty {
				cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))

				f, err := pty.Start(cmd)
				if err != nil {
					panic(err)
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
			} else {
				io.WriteString(sshSession, "No PTY requested.\n")
				sshSession.Exit(1)
			}
		},
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			ctx.SetValue("publickey", key)
			return true
		},
	}

	err = server.SetOption(ssh.HostKeyFile(path.Join(homeDir, ".ssh", "id_rsa")))
	if err != nil {
		panic(err)
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

func (s *SSHServer) Shutdown(reason string) {
	// Stop listening
}
