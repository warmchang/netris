package mino

import "code.rocketnine.space/tslocum/netris/pkg/event"

// Dark color ghosts are 60% original overlaid #777777
// Light color ghosts are 40% original overlaid #888888
var Colors = make([][]byte, 16)

type Block int

func (b Block) String() string {
	return string(b.Rune())
}

func (b Block) Rune() rune {
	switch b {
	case BlockNone:
		return ' '
	case BlockGhostJ, BlockGhostI, BlockGhostZ, BlockGhostO, BlockGhostT, BlockGhostS, BlockGhostL:
		return '▓'
	case BlockGarbage, BlockSolidJ, BlockSolidI, BlockSolidZ, BlockSolidO, BlockSolidT, BlockSolidS, BlockSolidL:
		return '█'
	default:
		return '?'
	}
}

const (
	BlockNone Block = iota
	BlockGarbage
	BlockGhostJ
	BlockGhostI
	BlockGhostZ
	BlockGhostO
	BlockGhostT
	BlockGhostS
	BlockGhostL
	BlockSolidJ
	BlockSolidI
	BlockSolidZ
	BlockSolidO
	BlockSolidT
	BlockSolidS
	BlockSolidL
)

var ColorToBlock = map[event.GameColor]Block{
	event.GameColorI:       BlockSolidI,
	event.GameColorO:       BlockSolidO,
	event.GameColorT:       BlockSolidT,
	event.GameColorJ:       BlockSolidJ,
	event.GameColorL:       BlockSolidL,
	event.GameColorS:       BlockSolidS,
	event.GameColorZ:       BlockSolidZ,
	event.GameColorIGhost:  BlockGhostI,
	event.GameColorOGhost:  BlockGhostO,
	event.GameColorTGhost:  BlockGhostT,
	event.GameColorJGhost:  BlockGhostJ,
	event.GameColorLGhost:  BlockGhostL,
	event.GameColorSGhost:  BlockGhostS,
	event.GameColorZGhost:  BlockGhostZ,
	event.GameColorGarbage: BlockGarbage,
}
