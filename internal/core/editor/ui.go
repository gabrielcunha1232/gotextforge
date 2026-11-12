package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

func (e *Editor) Frame() render.Frame {
	total := e.buf.LineCount()
	start, end := e.viewport.VisibleRange(total)
	vpHeight := e.viewport.Height

	if dirtyLine, ok := e.buf.TakeDirty(); ok {
		e.syntax.Invalidate(dirtyLine)
	}

	gw := e.gutterWidth()
	textWidth := max(e.width-gw, 0)
	lines := make([]render.StyledLine, 0, vpHeight)

	for i := start; i < end; i++ {
		current := i == e.cursor.Line
		line := e.gutter(i, gw, current)
		line = append(line, e.renderBufferLine(i, textWidth, current)...)
		if current {
			line = padTo(line, e.width, render.Style{Bg: colBgHighlight})
		}
		lines = append(lines, line)
	}
	for len(lines) < vpHeight {
		lines = append(lines, render.StyledLine{{Text: "~", Style: styleTilde}})
	}

	if e.showIntro() {
		e.overlayIntro(lines)
	}

	frame := render.Frame{
		Lines:       lines,
		Cursor:      e.screenCursor(start, gw),
		CursorShape: render.CursorBlock,
		Status:      e.statusLine(),
	}

	switch e.mode {
	case ModeInsert:
		frame.CursorShape = render.CursorBar
	case ModeCommand:
		cmd := append([]rune{':'}, e.cmdline...)
		frame.CursorShape = render.CursorBar
		frame.Cursor = render.ScreenPos{
			Row: vpHeight,
			Col: min(view.ScreenCol(cmd, len(cmd), view.DefaultTabStop), max(e.width-1, 0)),
		}
	}

	return frame
}

func (e *Editor) gutterWidth() int {
	digits := len(strconv.Itoa(e.buf.LineCount()))
	return max(digits, 3) + 2
}

func (e *Editor) gutter(i, width int, current bool) render.StyledLine {
	if width >= e.width {
		return nil
	}
	if current {
		return render.StyledLine{{Text: fmt.Sprintf(" %-*d ", width-2, i+1), Style: styleCursorLineNr}}
	}
	rel := i - e.cursor.Line
	if rel < 0 {
		rel = -rel
	}
	return render.StyledLine{{Text: fmt.Sprintf(" %*d ", width-2, rel), Style: styleLineNr}}
}

func (e *Editor) renderBufferLine(i, width int, current bool) render.StyledLine {
	runes := e.buf.Line(i)
	var bg render.Color
	if current {
		bg = colBgHighlight
	}

	if !e.highlight {
		plain := render.Style{Bg: bg}
		return view.RenderLine(runes, func(int) render.Style { return plain }, width, view.DefaultTabStop)
	}

	tokens, _ := syntax.HighlightLine(runes, e.syntax.StateAt(e.buf, i))
	styles := make([]render.Style, len(runes))
	for j := range styles {
		styles[j].Bg = bg
	}
	for _, t := range tokens {
		for j := t.Start; j < t.End; j++ {
			styles[j] = withBg(theme[t.Kind], bg)
		}
	}
	return view.RenderLine(runes, func(j int) render.Style { return styles[j] }, width, view.DefaultTabStop)
}

func padTo(line render.StyledLine, width int, style render.Style) render.StyledLine {
	if missing := width - line.Width(); missing > 0 {
		line = append(line, render.Span{Text: strings.Repeat(" ", missing), Style: style})
	}
	return line
}

func (e *Editor) screenCursor(top, gutter int) render.ScreenPos {
	col := gutter + view.ScreenCol(e.buf.Line(e.cursor.Line), e.cursor.Col, view.DefaultTabStop)
	if e.width > 0 && col > e.width-1 {
		col = e.width - 1
	}
	return render.ScreenPos{Row: e.cursor.Line - top, Col: col}
}

func sanitize(s string) string {
	return view.FormatLine([]rune(s), 1<<16, view.DefaultTabStop)
}

func (e *Editor) statusLine() render.StyledLine {
	if e.mode == ModeCommand {
		cmd := append([]rune{':'}, e.cmdline...)
		return render.StyledLine{{Text: view.FormatLine(cmd, e.width, view.DefaultTabStop)}}
	}

	modeStyle := render.Style{Fg: colBlack, Bg: modeColors[e.mode], Bold: true}

	name := sanitize(e.filename)
	if e.buf.Modified() {
		name += " [+]"
	}

	left := render.StyledLine{
		{Text: " " + modeNames[e.mode] + " ", Style: modeStyle},
		{Text: " " + name + " ", Style: styleStatusB},
	}
	if e.message != "" {
		style := styleStatusC
		if isErrorMessage(e.message) {
			style = styleStatusError
		}
		left = append(left, render.Span{Text: " " + sanitize(e.message), Style: style})
	}

	var right render.StyledLine
	if pending := e.parser.Pending(); pending != "" {
		right = append(right, render.Span{Text: " " + sanitize(pending) + " ", Style: styleStatusPending})
	}
	if e.highlight {
		right = append(right, render.Span{Text: " go ", Style: styleStatusC})
	}
	right = append(right,
		render.Span{Text: " " + e.progress() + " ", Style: styleStatusB},
		render.Span{Text: fmt.Sprintf(" %d:%d ", e.cursor.Line+1, e.cursor.Col+1), Style: modeStyle},
	)

	if left.Width()+right.Width() > e.width {
		right = nil
	}
	gap := e.width - left.Width() - right.Width()
	line := left
	if gap > 0 {
		line = append(line, render.Span{Text: strings.Repeat(" ", gap), Style: styleStatusC})
	}
	line = append(line, right...)
	return truncateLine(line, e.width)
}

func isErrorMessage(msg string) bool {
	return strings.HasPrefix(msg, "E") && len(msg) > 1 && msg[1] >= '0' && msg[1] <= '9' ||
		strings.HasPrefix(msg, "erro") || strings.Contains(msg, "bloqueado")
}

func (e *Editor) progress() string {
	total := e.buf.LineCount()
	switch {
	case total <= e.viewport.Height:
		return "All"
	case e.cursor.Line == 0:
		return "Top"
	case e.cursor.Line == total-1:
		return "Bot"
	default:
		return fmt.Sprintf("%d%%", (e.cursor.Line+1)*100/total)
	}
}

func truncateLine(line render.StyledLine, width int) render.StyledLine {
	if width <= 0 {
		return nil
	}
	out := make(render.StyledLine, 0, len(line))
	used := 0
	for _, s := range line {
		runes := []rune(s.Text)
		if used+len(runes) > width {
			runes = runes[:width-used]
		}
		if len(runes) > 0 {
			out = append(out, render.Span{Text: string(runes), Style: s.Style})
		}
		used += len(runes)
		if used >= width {
			break
		}
	}
	return out
}
