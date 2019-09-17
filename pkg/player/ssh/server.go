package ssh

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
	"unsafe"

	"git.sr.ht/~tslocum/netris/pkg/player"
	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

type SSHServer struct {
	NetrisPath string
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}
func (s *SSHServer) Host(newPlayers chan<- *player.ConnectingPlayer) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	server := &ssh.Server{
		Addr: "localhost:7777",
		Handler: func(sshSession ssh.Session) {
			cmd := exec.Command(s.NetrisPath)
			ptyReq, winCh, isPty := sshSession.Pty()
			if isPty {
				cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
				f, err := pty.Start(cmd)
				if err != nil {
					panic(err)
				}

				go func() {
					for win := range winCh {
						setWinsize(f, win.Width, win.Height)
					}
				}()

				go func() {
					io.Copy(f, sshSession)
				}()
				io.Copy(sshSession, f)
			} else {
				io.WriteString(sshSession, "No PTY requested.\n")
				sshSession.Exit(1)
			}
		},
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return true
		},
	}

	err = server.SetOption(ssh.HostKeyFile(path.Join(homeDir, ".ssh", "id_rsa")))
	if err != nil {
		panic(err)
	}

	go log.Fatal(server.ListenAndServe())
}

func (s *SSHServer) Shutdown(reason string) {
	// Stop listening
}