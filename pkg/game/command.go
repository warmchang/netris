package game

import (
	"strconv"
	"time"

	"gitlab.com/tslocum/netris/pkg/mino"
)

type Command int

// The order of these constants must be preserved
const (
	CommandUnknown Command = iota
	CommandDisconnect
	CommandPing
	CommandPong
	CommandNickname
	CommandMessage
	CommandNewGame
	CommandJoinGame
	CommandQuitGame
	CommandUpdateGame
	CommandStartGame
	CommandGameOver
	CommandUpdateMatrix
	CommandSendGarbage
	CommandReceiveGarbage
	CommandStats
	CommandListGames
)

func (c Command) String() string {
	switch c {
	case CommandUnknown:
		return "Unknown"
	case CommandDisconnect:
		return "Disconnect"
	case CommandPing:
		return "Ping"
	case CommandPong:
		return "Pong"
	case CommandNickname:
		return "Nickname"
	case CommandMessage:
		return "Message"
	case CommandNewGame:
		return "NewGame"
	case CommandJoinGame:
		return "JoinGame"
	case CommandQuitGame:
		return "QuitGame"
	case CommandUpdateGame:
		return "UpdateGame"
	case CommandStartGame:
		return "StartGame"
	case CommandGameOver:
		return "GameOver"
	case CommandUpdateMatrix:
		return "UpdateMatrix"
	case CommandSendGarbage:
		return "Garbage-OUT"
	case CommandReceiveGarbage:
		return "Garbage-IN"
	case CommandStats:
		return "Stats"
	case CommandListGames:
		return "ListGames"
	default:
		return strconv.Itoa(int(c))
	}
}

type GameCommandInterface interface {
	Command() Command
	Source() int
	SetSource(int)
}

type GameCommand struct {
	SourcePlayer int `json:"sp,omitempty"`
}

func (gc *GameCommand) Source() int {
	if gc == nil {
		return 0
	}

	return gc.SourcePlayer
}

func (gc *GameCommand) SetSource(source int) {
	if gc == nil {
		return
	}

	gc.SourcePlayer = source
}

type GameCommandDisconnect struct {
	GameCommand
	Player  int    `json:"p,omitempty"`
	Message string `json:"m,omitempty"`
}

func (gc GameCommandDisconnect) Command() Command {
	return CommandDisconnect
}

type GameCommandPing struct {
	GameCommand
	Message string `json:"m,omitempty"`
}

func (gc GameCommandPing) Command() Command {
	return CommandPing
}

type GameCommandPong struct {
	GameCommand
	Message string `json:"m,omitempty"`
}

func (gc GameCommandPong) Command() Command {
	return CommandPong
}

type GameCommandNickname struct {
	GameCommand
	Player   int    `json:"p,omitempty"`
	Nickname string `json:"n,omitempty"`
}

func (gc GameCommandNickname) Command() Command {
	return CommandNickname
}

type GameCommandMessage struct {
	GameCommand
	Player  int    `json:"p,omitempty"`
	Message string `json:"m,omitempty"`
}

func (gc GameCommandMessage) Command() Command {
	return CommandMessage
}

type GameCommandJoinGame struct {
	GameCommand
	Version  int    `json:"v,omitempty"`
	Name     string `json:"n,omitempty"`
	GameID   int    `json:"g,omitempty"`
	PlayerID int    `json:"p,omitempty"`

	Listing ListedGame `json:"l,omitempty"`
}

func (gc GameCommandJoinGame) Command() Command {
	return CommandJoinGame
}

type GameCommandQuitGame struct {
	GameCommand
	Player int `json:"p,omitempty"`
}

func (gc GameCommandQuitGame) Command() Command {
	return CommandQuitGame
}

type GameCommandUpdateGame struct {
	GameCommand
	Players map[int]string `json:"p,omitempty"`
}

func (gc GameCommandUpdateGame) Command() Command {
	return CommandUpdateGame
}

type GameCommandStartGame struct {
	GameCommand
	Seed    int64 `json:"s,omitempty"`
	Started bool  `json:"st,omitempty"`
}

func (gc GameCommandStartGame) Command() Command {
	return CommandStartGame
}

type GameCommandUpdateMatrix struct {
	GameCommand
	Matrixes map[int]*mino.Matrix `json:"m,omitempty"`
}

func (gc GameCommandUpdateMatrix) Command() Command {
	return CommandUpdateMatrix
}

type GameCommandGameOver struct {
	GameCommand
	Player int    `json:"p,omitempty"`
	Winner string `json:"w,omitempty"`
}

func (gc GameCommandGameOver) Command() Command {
	return CommandGameOver
}

type GameCommandSendGarbage struct {
	GameCommand
	Lines int `json:"l,omitempty"`
}

func (gc GameCommandSendGarbage) Command() Command {
	return CommandSendGarbage
}

type GameCommandReceiveGarbage struct {
	GameCommand
	Lines int `json:"l,omitempty"`
}

func (gc GameCommandReceiveGarbage) Command() Command {
	return CommandReceiveGarbage
}

type GameCommandStats struct {
	GameCommand
	Created time.Time `json:"c,omitempty"`
	Players int       `json:"p,omitempty"`
	Games   int       `json:"g,omitempty"`
}

func (gc GameCommandStats) Command() Command {
	return CommandStats
}

type ListedGame struct {
	ID         int
	Name       string `json:"n,omitempty"`
	Players    int    `json:"p,omitempty"`
	MaxPlayers int    `json:"pl,omitempty"`
	SpeedLimit int    `json:"sl,omitempty"`
}
type GameCommandListGames struct {
	GameCommand

	Games []*ListedGame `json:"g,omitempty"`
}

func (gc GameCommandListGames) Command() Command {
	return CommandListGames
}
