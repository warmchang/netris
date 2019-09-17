package mino

type Block int

func (b Block) String() string {
	return string(b.Rune())
}

func (b Block) Rune() rune {
	switch b {
	case BlockNone:
		return ' '
	case BlockGhost:
		return '▒'
	case BlockSolidBlue, BlockSolidCyan, BlockSolidRed, BlockSolidYellow, BlockSolidMagenta, BlockSolidGreen, BlockSolidOrange:
		return '█'
	default:
		return '?'
	}
}

const (
	BlockNone Block = iota
	BlockGhost
	BlockGarbage
	BlockSolidBlue
	BlockSolidCyan
	BlockSolidRed
	BlockSolidYellow
	BlockSolidMagenta
	BlockSolidGreen
	BlockSolidOrange
)
