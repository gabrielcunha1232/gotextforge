package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
)

var (
	colBlack       = render.RGB(0x0b, 0x0e, 0x13)
	colBgHighlight = render.RGB(0x1a, 0x22, 0x2e)
	colBgStatus    = render.RGB(0x16, 0x1b, 0x22)
	colFg          = render.RGB(0xd6, 0xe2, 0xec)
	colFgDark      = render.RGB(0xaa, 0xb7, 0xc4)
	colFgGutter    = render.RGB(0x3a, 0x46, 0x54)
	colDark3       = render.RGB(0x5a, 0x6b, 0x7a)
	colComment     = render.RGB(0x5f, 0x70, 0x83)
	colBlue        = render.RGB(0x00, 0xad, 0xd8)
	colBlue1       = render.RGB(0x38, 0xbd, 0xd0)
	colMagenta     = render.RGB(0xb0, 0x8c, 0xf0)
	colOrange      = render.RGB(0xe8, 0xa2, 0x4a)
	colYellow      = render.RGB(0xe6, 0xc0, 0x72)
	colGreen       = render.RGB(0x5b, 0xc9, 0x8a)
	colRed         = render.RGB(0xff, 0x6b, 0x7a)
	colTeal        = render.RGB(0x2b, 0xb3, 0xa3)
	colBgFloat     = render.RGB(0x0e, 0x11, 0x16)
	colBgSel       = render.RGB(0x20, 0x30, 0x40)
	colBorder      = render.RGB(0x2b, 0x7a, 0x99)
)

var theme = [...]render.Style{
	syntax.KindPlain:       {},
	syntax.KindKeyword:     {Fg: colMagenta, Italic: true},
	syntax.KindString:      {Fg: colGreen},
	syntax.KindNumber:      {Fg: colOrange},
	syntax.KindComment:     {Fg: colComment, Italic: true},
	syntax.KindPredeclared: {Fg: colBlue1},
	syntax.KindFunc:        {Fg: colBlue},
}

var (
	styleLineNr        = render.Style{Fg: colFgGutter}
	styleCursorLineNr  = render.Style{Fg: colOrange, Bold: true, Bg: colBgHighlight}
	styleTilde         = render.Style{Fg: colDark3}
	styleStatusB       = render.Style{Fg: colFg, Bg: colFgGutter}
	styleStatusC       = render.Style{Fg: colFgDark, Bg: colBgStatus}
	styleStatusError   = render.Style{Fg: colRed, Bg: colBgStatus, Bold: true}
	styleStatusPending = render.Style{Fg: colOrange, Bg: colBgStatus, Bold: true}
	styleIntroSub      = render.Style{Fg: colComment, Italic: true}
	styleSearch        = render.Style{Fg: colFg, Bg: render.RGB(0x3d, 0x59, 0xa1)}
	styleSearchCurrent = render.Style{Fg: colBlack, Bg: colOrange, Bold: true}
	styleIntroKey      = render.Style{Fg: colBlue, Bold: true}
	styleIntroDesc     = render.Style{Fg: colFgDark}
)

var modeColors = map[Mode]render.Color{
	ModeNormal:   colBlue,
	ModeInsert:   colGreen,
	ModeCommand:  colYellow,
	ModeSearch:   colYellow,
	ModePrompt:   colYellow,
	ModeExplorer: colTeal,
	ModeFinder:   colMagenta,
	ModeTerminal: colGreen,
	ModeDocker:   colBlue1,
}

var modeNames = map[Mode]string{
	ModeNormal:   "NORMAL",
	ModeInsert:   "INSERT",
	ModeCommand:  "COMMAND",
	ModeSearch:   "SEARCH",
	ModePrompt:   "PROMPT",
	ModeExplorer: "ARQUIVOS",
	ModeFinder:   "BUSCAR",
	ModeTerminal: "TERMINAL",
	ModeDocker:   "DOCKER",
}

func withBg(s render.Style, bg render.Color) render.Style {
	s.Bg = bg
	return s
}
