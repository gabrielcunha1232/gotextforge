package editor

import (
	"regexp"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

const (
	hoverMaxWidth  = 80
	hoverMaxHeight = 15
)

var markdownLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

type hoverLine struct {
	text      string
	code      bool
	separator bool
	heading   bool
}

type hoverPopup struct {
	lines []hoverLine
}

func newHoverPopup(markdown string) *hoverPopup {
	var lines []hoverLine
	inCode := false
	for _, raw := range strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(raw)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			lines = append(lines, hoverLine{text: raw, code: true})
			continue
		}
		if trimmed == "---" {
			lines = append(lines, hoverLine{separator: true})
			continue
		}
		text := markdownLink.ReplaceAllString(raw, "$1")
		text = strings.ReplaceAll(text, "`", "")
		heading := strings.HasPrefix(trimmed, "#")
		if heading {
			text = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		}
		lines = append(lines, hoverLine{text: strings.TrimRight(text, " "), heading: heading})
	}

	var cleaned []hoverLine
	for _, l := range lines {
		blank := !l.code && !l.separator && l.text == ""
		if blank && (len(cleaned) == 0 || isBlankHover(cleaned[len(cleaned)-1])) {
			continue
		}
		cleaned = append(cleaned, l)
	}
	for len(cleaned) > 0 && (isBlankHover(cleaned[len(cleaned)-1]) || cleaned[len(cleaned)-1].separator) {
		cleaned = cleaned[:len(cleaned)-1]
	}
	if len(cleaned) == 0 {
		return nil
	}
	return &hoverPopup{lines: cleaned}
}

func isBlankHover(l hoverLine) bool {
	return !l.code && !l.separator && l.text == ""
}

func wrapWords(text string, width int) []string {
	if width <= 0 {
		return nil
	}
	var out []string
	var cur []rune
	for _, word := range strings.Fields(text) {
		w := []rune(word)
		switch {
		case len(cur) == 0:
			cur = w
		case len(cur)+1+len(w) <= width:
			cur = append(append(cur, ' '), w...)
		default:
			out = append(out, string(cur))
			cur = w
		}
		for len(cur) > width {
			out = append(out, string(cur[:width]))
			cur = cur[width:]
		}
	}
	if len(cur) > 0 || len(out) == 0 {
		out = append(out, string(cur))
	}
	return out
}

func (h *hoverPopup) render(maxInner int) []render.StyledLine {
	inner := 0
	for _, l := range h.lines {
		if !l.separator {
			inner = max(inner, len([]rune(view.FormatLine([]rune(l.text), 1<<16, view.DefaultTabStop))))
		}
	}
	inner = max(min(inner, maxInner, hoverMaxWidth), 1)

	textStyle := render.Style{Fg: colFg, Bg: colBgFloat}
	var body []render.StyledLine
	var state syntax.State
	for _, l := range h.lines {
		switch {
		case l.separator:
			body = append(body, render.StyledLine{{Text: strings.Repeat("─", inner), Style: render.Style{Fg: colFgGutter, Bg: colBgFloat}}})
		case l.code:
			runes := []rune(l.text)
			tokens, out := syntax.HighlightLine(runes, state)
			state = out
			styles := make([]render.Style, len(runes))
			for j := range styles {
				styles[j] = textStyle
			}
			for _, t := range tokens {
				for j := t.Start; j < t.End; j++ {
					styles[j] = withBg(theme[t.Kind], colBgFloat)
					if styles[j].Fg == (render.Color{}) {
						styles[j].Fg = colFg
					}
				}
			}
			line := view.RenderLine(runes, func(j int) render.Style { return styles[j] }, inner, view.DefaultTabStop)
			body = append(body, line)
		default:
			style := textStyle
			if l.heading {
				style = render.Style{Fg: colBlue, Bg: colBgFloat, Bold: true}
			}
			for _, w := range wrapWords(l.text, inner) {
				body = append(body, render.StyledLine{{Text: sanitize(w), Style: style}})
			}
		}
	}
	if len(body) > hoverMaxHeight {
		body = append(body[:hoverMaxHeight-1], render.StyledLine{{Text: "…", Style: render.Style{Fg: colComment, Bg: colBgFloat}}})
	}

	border := render.Style{Fg: colBorder, Bg: colBgFloat}
	fill := render.Style{Bg: colBgFloat}
	box := []render.StyledLine{{{Text: "╭" + strings.Repeat("─", inner+2) + "╮", Style: border}}}
	for _, l := range body {
		row := render.StyledLine{{Text: "│ ", Style: border}}
		row = append(row, truncateLine(l, inner)...)
		row = padTo(row, inner+2, fill)
		row = append(row, render.Span{Text: " │", Style: border})
		box = append(box, row)
	}
	box = append(box, render.StyledLine{{Text: "╰" + strings.Repeat("─", inner+2) + "╯", Style: border}})
	return box
}

func dropCells(line render.StyledLine, n int) render.StyledLine {
	var out render.StyledLine
	for _, s := range line {
		runes := []rune(s.Text)
		if n >= len(runes) {
			n -= len(runes)
			continue
		}
		out = append(out, render.Span{Text: string(runes[n:]), Style: s.Style})
		n = 0
	}
	return out
}

func overlay(line render.StyledLine, col int, insert render.StyledLine, width int) render.StyledLine {
	head := padTo(truncateLine(line, col), col, render.Style{})
	out := append(render.StyledLine{}, head...)
	out = append(out, insert...)
	out = append(out, dropCells(line, col+insert.Width())...)
	return truncateLine(out, width)
}

func (e *Editor) placeBox(lines []render.StyledLine, box []render.StyledLine, anchorRow, anchorCol int) {
	if len(box) == 0 || len(lines) == 0 {
		return
	}
	boxW := box[0].Width()
	below := len(lines) - anchorRow - 1
	above := anchorRow
	var row int
	switch {
	case len(box) <= below:
		row = anchorRow + 1
	case len(box) <= above:
		row = anchorRow - len(box)
	case below >= above:
		box = shrinkBox(box, below)
		row = anchorRow + 1
	default:
		box = shrinkBox(box, above)
		row = anchorRow - len(box)
	}
	if len(box) == 0 {
		return
	}
	col := anchorCol
	if col+boxW > e.width {
		col = max(e.width-boxW, 0)
	}
	for i, b := range box {
		if row+i < len(lines) {
			lines[row+i] = overlay(lines[row+i], col, b, e.width)
		}
	}
}

func shrinkBox(box []render.StyledLine, n int) []render.StyledLine {
	switch {
	case n <= 0:
		return nil
	case n >= len(box):
		return box
	case n < 3:
		return box[:n]
	}
	border := render.Style{Fg: colBorder, Bg: colBgFloat}
	inner := box[0].Width() - 4
	more := render.StyledLine{
		{Text: "│ ", Style: border},
		{Text: padRight("…", inner), Style: render.Style{Fg: colComment, Bg: colBgFloat}},
		{Text: " │", Style: border},
	}
	out := append([]render.StyledLine{}, box[:n-2]...)
	return append(out, more, box[len(box)-1])
}
