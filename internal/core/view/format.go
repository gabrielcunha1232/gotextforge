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
	return (r >= 0 && r < 0x20 && r != '\t') || r == 0x7f
}

func controlNotation(r rune) [2]rune {
	if r == 0x7f {
		return [2]rune{'^', '?'}
	}
	return [2]rune{'^', r + 64}
}
