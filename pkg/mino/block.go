package mino

type Block int

const (
	BlockNone Block = iota
	BlockGhost
	BlockGarbage
	BlockSolid
)
