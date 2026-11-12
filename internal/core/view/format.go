package view

import (
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

const DefaultTabStop = 4

func FormatLine(runes []rune, width, tabStop int) string {
	var sb strings.Builder
	layout(runes, width, tabStop, func(_ int, r rune) {
		sb.WriteRune(r)
	})
	return sb.String()
}

func RenderLine(runes []rune, styleAt func(i int) render.Style, width, tabStop int) render.StyledLine {
	var line render.StyledLine
	var sb strings.Builder
	var current render.Style
	open := false

	layout(runes, width, tabStop, func(src int, r rune) {
		style := styleAt(src)
		if open && style != current {
			line = append(line, render.Span{Text: sb.String(), Style: current})
			sb.Reset()
		}
		current = style
		open = true
		sb.WriteRune(r)
	})

	if open {
		line = append(line, render.Span{Text: sb.String(), Style: current})
	}
	return line
}

func layout(runes []rune, width, tabStop int, emit func(src int, r rune)) {
	if width <= 0 {
		return
	}
	if tabStop <= 0 {
		tabStop = DefaultTabStop
	}

	col := 0
	for i, r := range runes {
		if col >= width {
			return
		}

		w := cellWidth(r, col, tabStop)
		switch {
		case r == '\t':
			for s := 0; s < w && col < width; s++ {
				emit(i, ' ')
				col++
			}
		case isControlRune(r):
			for _, cr := range controlNotation(r) {
				if col >= width {
					return
				}
				emit(i, cr)
				col++
			}
		default:
			emit(i, r)
			col++
		}
	}
}

func ScreenCol(runes []rune, runeCol, tabStop int) int {
	if tabStop <= 0 {
		tabStop = DefaultTabStop
	}
	col := 0
	for i, r := range runes {
		if i >= runeCol {
			break
		}
		col += cellWidth(r, col, tabStop)
	}
	return col
}

func cellWidth(r rune, col, tabStop int) int {
	switch {
	case r == '\t':
		return tabStop - (col % tabStop)
	case isControlRune(r):
		return len(controlNotation(r))
	default:
		return 1
	}
}

func isControlRune(r rune) bool {
	return (r >= 0 && r < 0x20 && r != '\t') || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

func controlNotation(r rune) []rune {
	switch {
	case r == 0x7f:
		return []rune{'^', '?'}
	case r >= 0x80:
		const hex = "0123456789abcdef"
		return []rune{'<', rune(hex[r>>4]), rune(hex[r&0xf]), '>'}
	default:
		return []rune{'^', r + 64}
	}
}
