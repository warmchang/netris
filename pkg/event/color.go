package event

type GameColor string

const (
	GameColorI       = "i"
	GameColorO       = "o"
	GameColorT       = "t"
	GameColorJ       = "j"
	GameColorL       = "l"
	GameColorS       = "s"
	GameColorZ       = "z"
	GameColorIGhost  = "i-ghost"
	GameColorOGhost  = "o-ghost"
	GameColorTGhost  = "t-ghost"
	GameColorJGhost  = "j-ghost"
	GameColorLGhost  = "l-ghost"
	GameColorSGhost  = "s-ghost"
	GameColorZGhost  = "z-ghost"
	GameColorGarbage = "garbage"
	GameColorBorder  = "border"
)

var DefaultColors = map[GameColor]string{
	GameColorJ:       "#2864ff",
	GameColorI:       "#00eeee",
	GameColorZ:       "#ee0000",
	GameColorO:       "#dddd00",
	GameColorT:       "#c000cc",
	GameColorS:       "#00e900",
	GameColorL:       "#ff7308",
	GameColorJGhost:  "#6e7bc3",
	GameColorIGhost:  "#6bbaba",
	GameColorZGhost:  "#ba6b6b",
	GameColorOGhost:  "#b1b16b",
	GameColorTGhost:  "#a16ba8",
	GameColorSGhost:  "#6bb76b",
	GameColorLGhost:  "#c3806c",
	GameColorGarbage: "#999999",
	GameColorBorder:  "#444444",
}
