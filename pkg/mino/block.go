package mino

type Block int

func (b Block) String() string {
	return string(BlockToRune(b))
}

func BlockToRune(block Block) rune {
	switch block {
	case BlockNone:
		return ' '
	case BlockGhost:
		return '▒'
	case BlockSolid:
		return '█'
	default:
		return '?'
	}
}

const (
	BlockNone Block = iota
	BlockGhost
	BlockGarbage
	BlockSolid
)
