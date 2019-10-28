package mino

// Dark color ghosts are 60% original overlaid #777777
// Light color ghosts are 40% original overlaid #888888
var Colors = [][]byte{
	BlockNone:         []byte("#000000"),
	BlockGarbage:      []byte("#999999"),
	BlockGhostBlue:    []byte("#6e7bc3"),
	BlockGhostCyan:    []byte("#6bbaba"),
	BlockGhostRed:     []byte("#ba6b6b"),
	BlockGhostYellow:  []byte("#b1b16b"),
	BlockGhostMagenta: []byte("#a16ba8"),
	BlockGhostGreen:   []byte("#6bb76b"),
	BlockGhostOrange:  []byte("#c3806c"),
	BlockSolidBlue:    []byte("#2864ff"),
	BlockSolidCyan:    []byte("#00eeee"),
	BlockSolidRed:     []byte("#ee0000"),
	BlockSolidYellow:  []byte("#dddd00"),
	BlockSolidMagenta: []byte("#c000cc"),
	BlockSolidGreen:   []byte("#00e900"),
	BlockSolidOrange:  []byte("#ff7308"),
}

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
