package game

import (
	"strconv"
	"time"

	"git.sr.ht/~tslocum/netris/pkg/mino"
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
	SourcePlayer int
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
	Player  int
	Message string
}

func (gc GameCommandDisconnect) Command() Command {
	return CommandDisconnect
}

type GameCommandPing struct {
	GameCommand
	Message string
}

func (gc GameCommandPing) Command() Command {
	return CommandPing
}

type GameCommandPong struct {
	GameCommand
	Message string
}

func (gc GameCommandPong) Command() Command {
	return CommandPong
}

type GameCommandNickname struct {
	GameCommand
	Player   int
	Nickname string
}

func (gc GameCommandNickname) Command() Command {
	return CommandNickname
}

type GameCommandMessage struct {
	GameCommand
	Player  int
	Message string
}

func (gc GameCommandMessage) Command() Command {
	return CommandMessage
}

type GameCommandJoinGame struct {
	GameCommand
	Version  int
	Name     string
	GameID   int
	PlayerID int
}

func (gc GameCommandJoinGame) Command() Command {
	return CommandJoinGame
}

type GameCommandQuitGame struct {
	GameCommand
	Player int
}

func (gc GameCommandQuitGame) Command() Command {
	return CommandQuitGame
}

type GameCommandUpdateGame struct {
	GameCommand
	Players map[int]string
}

func (gc GameCommandUpdateGame) Command() Command {
	return CommandUpdateGame
}

type GameCommandStartGame struct {
	GameCommand
	Seed    int64
	Started bool
}

func (gc GameCommandStartGame) Command() Command {
	return CommandStartGame
}

type GameCommandUpdateMatrix struct {
	GameCommand
	Matrixes map[int]*mino.Matrix
}

func (gc GameCommandUpdateMatrix) Command() Command {
	return CommandUpdateMatrix
}

type GameCommandGameOver struct {
	GameCommand
	Player int
	Winner string
}

func (gc GameCommandGameOver) Command() Command {
	return CommandGameOver
}

type GameCommandSendGarbage struct {
	GameCommand
	Lines int
}

func (gc GameCommandSendGarbage) Command() Command {
	return CommandSendGarbage
}

type GameCommandReceiveGarbage struct {
	GameCommand
	Lines int
}

func (gc GameCommandReceiveGarbage) Command() Command {
	return CommandReceiveGarbage
}

type GameCommandStats struct {
	GameCommand
	Created time.Time
	Players int
	Games   int
}

func (gc GameCommandStats) Command() Command {
	return CommandStats
}
