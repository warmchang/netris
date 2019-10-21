package game

import (
	"regexp"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/mino"
)

const (
	CommandQueueSize = 10
	LogQueueSize     = 10
	PlayerHost       = -1
	PlayerUnknown    = 0
)

var nickRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-!@#$%^&*+=,./]+`)

type ConnectingPlayer struct {
	Name string
}

type Player struct {
	Name string

	*Conn

	Score   int
	Preview *mino.Matrix
	Matrix  *mino.Matrix
	Moved   time.Time     // Time of last piece move
	Idle    time.Duration // Time spent idling

	pendingGarbage       int
	totalGarbageSent     int
	totalGarbageReceived int
}

func NewPlayer(name string, conn *Conn) *Player {
	if conn == nil {
		conn = &Conn{}
	}

	p := &Player{Name: Nickname(name), Conn: conn, Moved: time.Now()}

	return p
}

func Nickname(nick string) string {
	nick = nickRegexp.ReplaceAllString(nick, "")
	if len(nick) > 10 {
		nick = nick[:10]
	} else if nick == "" {
		nick = "Anonymous"
	}

	return nick
}
