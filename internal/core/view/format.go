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

		if r == '\t' {
			spaces := tabStop - (col % tabStop)
			for s := 0; s < spaces && col < width; s++ {
				out = append(out, ' ')
				col++
			}
			continue
		}

		if isControlRune(r) {
			ctrl := controlNotation(r)
			for _, cr := range ctrl {
				if col >= width {
					break
				}
				out = append(out, cr)
				col++
			}
			continue
		}

		out = append(out, r)
		col++
	}

	return string(out)
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
