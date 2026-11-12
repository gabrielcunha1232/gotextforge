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

		out = append(out, r)
		col++
	}

	return string(out)
}
