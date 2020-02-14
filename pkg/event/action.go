package event

type GameAction string

const (
	ActionUnknown   = ""
	ActionRotateCCW = "rotate-ccw"
	ActionRotateCW  = "rotate-cw"
	ActionMoveLeft  = "move-left"
	ActionMoveRight = "move-right"
	ActionSoftDrop  = "soft-drop"
	ActionHardDrop  = "hard-drop"
	ActionPing      = "ping"
	ActionStats     = "stats"
	ActionNick      = "nick"
)
