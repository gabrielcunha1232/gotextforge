package view

const DefaultTabStop = 4

func FormatLine(runes []rune, width, tabStop int) string {
	if width <= 0 {
		return ""
	}
	if tabStop <= 0 {
		tabStop = DefaultTabStop
	}

	out := make([]rune, 0, width)
	col := 0

	for _, r := range runes {
		if col >= width {
			break
		}

		w := cellWidth(r, col, tabStop)
		switch {
		case r == '\t':
			for s := 0; s < w && col < width; s++ {
				out = append(out, ' ')
				col++
			}
		case isControlRune(r):
			for _, cr := range controlNotation(r) {
				if col >= width {
					break
				}
				out = append(out, cr)
				col++
			}
		default:
			out = append(out, r)
			col++
		}
	}

	return string(out)
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
