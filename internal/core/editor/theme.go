package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
)

var (
	colBlack       = render.RGB(0x15, 0x16, 0x1e)
	colBgHighlight = render.RGB(0x29, 0x2e, 0x42)
	colBgStatus    = render.RGB(0x1f, 0x23, 0x35)
	colFg          = render.RGB(0xc0, 0xca, 0xf5)
	colFgDark      = render.RGB(0xa9, 0xb1, 0xd6)
	colFgGutter    = render.RGB(0x3b, 0x42, 0x61)
	colDark3       = render.RGB(0x54, 0x5c, 0x7e)
	colComment     = render.RGB(0x56, 0x5f, 0x89)
	colBlue        = render.RGB(0x7a, 0xa2, 0xf7)
	colBlue1       = render.RGB(0x2a, 0xc3, 0xde)
	colMagenta     = render.RGB(0xbb, 0x9a, 0xf7)
	colOrange      = render.RGB(0xff, 0x9e, 0x64)
	colYellow      = render.RGB(0xe0, 0xaf, 0x68)
	colGreen       = render.RGB(0x9e, 0xce, 0x6a)
	colRed         = render.RGB(0xf7, 0x76, 0x8e)
	colTeal        = render.RGB(0x1a, 0xbc, 0x9c)
	colBgFloat     = render.RGB(0x16, 0x16, 0x1e)
	colBgSel       = render.RGB(0x28, 0x34, 0x57)
	colBorder      = render.RGB(0x27, 0xa1, 0xb9)
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
	ModeNormal:  colBlue,
	ModeInsert:  colGreen,
	ModeCommand: colYellow,
	ModeSearch:  colYellow,
}

var modeNames = map[Mode]string{
	ModeNormal:  "NORMAL",
	ModeInsert:  "INSERT",
	ModeCommand: "COMMAND",
	ModeSearch:  "SEARCH",
}

func withBg(s render.Style, bg render.Color) render.Style {
	s.Bg = bg
	return s
}
