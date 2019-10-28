package mino

type Block int

func (b Block) String() string {
	return string(b.Rune())
}

func (b Block) Rune() rune {
	switch b {
	case BlockNone:
		return ' '
	case BlockGhostBlue, BlockGhostCyan, BlockGhostRed, BlockGhostYellow, BlockGhostMagenta, BlockGhostGreen, BlockGhostOrange:
		return '▓'
	case BlockGarbage, BlockSolidBlue, BlockSolidCyan, BlockSolidRed, BlockSolidYellow, BlockSolidMagenta, BlockSolidGreen, BlockSolidOrange:
		return '█'
	default:
		return '?'
	}
}

func (b Block) Color() []byte {
	// Dark color ghosts are 60% original overlaid #777777
	// Light color ghosts are 40% original overlaid #888888

	switch b {
	case BlockNone:
		return []byte("#000000")
	case BlockGhostBlue:
		return []byte("#6e7bc3")
	case BlockSolidBlue:
		return []byte("#2864ff")
	case BlockGhostCyan:
		return []byte("#6bbaba")
	case BlockSolidCyan:
		return []byte("#00eeee")
	case BlockGhostRed:
		return []byte("#ba6b6b")
	case BlockSolidRed:
		return []byte("#ee0000")
	case BlockGhostYellow:
		return []byte("#b1b16b")
	case BlockSolidYellow:
		return []byte("#dddd00")
	case BlockGhostMagenta:
		return []byte("#a16ba8")
	case BlockSolidMagenta:
		return []byte("#c000cc")
	case BlockGhostGreen:
		return []byte("#6bb76b")
	case BlockSolidGreen:
		return []byte("#00e900")
	case BlockGhostOrange:
		return []byte("#c3806c")
	case BlockSolidOrange:
		return []byte("#ff7308")
	case BlockGarbage:
		return []byte("#999999")
	default:
		return []byte("#ffffff")
	}
}

const (
	BlockNone Block = iota
	BlockGarbage
	BlockGhostBlue
	BlockGhostCyan
	BlockGhostRed
	BlockGhostYellow
	BlockGhostMagenta
	BlockGhostGreen
	BlockGhostOrange
	BlockSolidBlue
	BlockSolidCyan
	BlockSolidRed
	BlockSolidYellow
	BlockSolidMagenta
	BlockSolidGreen
	BlockSolidOrange
)
