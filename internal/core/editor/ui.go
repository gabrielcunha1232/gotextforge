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
	fullWidth := e.width
	panel := 0
	if e.explorer.open && e.explorer.tree != nil {
		panel = explorerWidth(fullWidth)
		if panel == 0 && e.mode == ModeExplorer {
			panel = fullWidth
		}
	}
	e.width = fullWidth - panel
	frame := e.editorFrame()
	e.width = fullWidth

	if panel > 0 {
		side := e.explorerPanel(len(frame.Lines), panel)
		for i := range frame.Lines {
			if panel == fullWidth {
				frame.Lines[i] = side[i]
			} else {
				frame.Lines[i] = append(append(render.StyledLine{}, side[i]...), frame.Lines[i]...)
			}
		}
		if e.mode != ModeCommand && e.mode != ModeSearch && e.mode != ModePrompt {
			frame.Cursor.Col += panel
		}
	}

	if rows := e.bottomRows(); rows > 0 {
		if e.dock.open {
			frame.Lines = append(frame.Lines, e.dockerPanel(rows, fullWidth)...)
		} else {
			frame.Lines = append(frame.Lines, e.terminalPanel(rows, fullWidth)...)
		}
		switch e.mode {
		case ModeCommand, ModeSearch, ModePrompt:
			frame.Cursor.Row = len(frame.Lines)
		}
	}

	if e.mode == ModeFinder && e.finder != nil {
		box, row, col := e.finderBox(len(frame.Lines))
		for i, b := range box {
			if row+i < len(frame.Lines) {
				frame.Lines[row+i] = overlay(frame.Lines[row+i], col, b, fullWidth)
			}
		}
		frame.Cursor = render.ScreenPos{Row: row + 1, Col: min(col+4+len([]rune(sanitize(string(e.finder.query)))), fullWidth-1)}
		frame.CursorShape = render.CursorBar
	}
	if e.mode == ModeExplorer || e.mode == ModeTerminal || e.mode == ModeDocker {
		frame.HideCursor = true
	}
	frame.Status = e.statusLine()
	return frame
}

func (e *Editor) editorFrame() render.Frame {
	total := e.buf.LineCount()
	start, end := e.viewport.VisibleRange(total)
	vpHeight := e.viewport.Height

	if dirtyLine, ok := e.buf.TakeDirty(); ok {
		e.syntax.Invalidate(dirtyLine)
	}

	gw := e.gutterWidth()
	textWidth := max(e.width-gw, 0)
	lines := make([]render.StyledLine, 0, vpHeight)

	diags := e.diagnosticsByLine(start, end-1)
	for i := start; i < end; i++ {
		current := i == e.cursor.Line
		line := e.gutter(i, gw, current, diags[i])
		line = append(line, e.renderBufferLine(i, textWidth, current, diags[i])...)
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

	cursor := e.screenCursor(start, gw)
	switch {
	case e.hover != nil && e.mode == ModeNormal:
		e.placeBox(lines, e.hover.render(max(e.width-6, 1)), cursor.Row, cursor.Col)
	case e.comp != nil && e.mode == ModeInsert:
		anchor := gw + view.ScreenCol(e.buf.Line(e.cursor.Line), e.comp.start.Col, view.DefaultTabStop)
		e.placeBox(lines, e.completionBox(), cursor.Row, max(anchor-7, 0))
	}

	if tabs := e.tabline(); tabs != nil {
		lines = append([]render.StyledLine{tabs}, lines...)
		cursor.Row++
	}

	frame := render.Frame{
		Lines:       lines,
		Cursor:      cursor,
		CursorShape: render.CursorBlock,
		Status:      e.statusLine(),
	}

	switch e.mode {
	case ModeInsert:
		frame.CursorShape = render.CursorBar
	case ModeCommand, ModeSearch, ModePrompt:
		cmd := e.promptLine()
		frame.CursorShape = render.CursorBar
		frame.Cursor = render.ScreenPos{
			Row: len(lines),
			Col: min(view.ScreenCol(cmd, len(cmd), view.DefaultTabStop), max(e.width-1, 0)),
		}
	}

	return frame
}

func (e *Editor) gutterWidth() int {
	digits := len(strconv.Itoa(e.buf.LineCount()))
	return max(digits, 3) + 2
}

func (e *Editor) gutter(i, width int, current bool, diags []lineDiag) render.StyledLine {
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
	style := styleLineNr
	if len(diags) > 0 {
		style = render.Style{Fg: severityColor(worstSeverity(diags)), Bold: true}
	}
	return render.StyledLine{{Text: fmt.Sprintf(" %*d ", width-2, rel), Style: style}}
}

func (e *Editor) renderBufferLine(i, width int, current bool, diags []lineDiag) render.StyledLine {
	runes := e.buf.Line(i)
	var bg render.Color
	if current {
		bg = colBgHighlight
	}

	styles := make([]render.Style, len(runes))
	for j := range styles {
		styles[j].Bg = bg
	}
	if e.highlight {
		tokens, _ := syntax.HighlightLine(runes, e.syntax.StateAt(e.buf, i))
		for _, t := range tokens {
			for j := t.Start; j < t.End; j++ {
				styles[j] = withBg(theme[t.Kind], bg)
			}
		}
	}
	e.applySearchHighlight(i, runes, styles)
	applyDiagnostics(styles, diags)

	line := view.RenderLine(runes, func(j int) render.Style { return styles[j] }, width, view.DefaultTabStop)
	if text, sev := virtualText(diags); text != "" {
		if room := width - line.Width(); room > 4 {
			line = append(line, render.Span{Text: truncateText(text, room), Style: render.Style{Fg: severityColor(sev), Bg: bg, Italic: true}})
		}
	}
	return line
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
	if e.mode == ModeCommand || e.mode == ModeSearch || e.mode == ModePrompt {
		cmd := e.promptLine()
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
	right = append(right, e.lspStatusSpans()...)
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
