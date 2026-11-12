package editor

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

type Editor struct {
	buf      *buffer.Buffer
	viewport view.Viewport
	cursor   buffer.Pos
	wantCol  int
	filename string
	width    int
	height   int
}

func New(buf *buffer.Buffer, filename string) *Editor {
	if buf == nil {
		buf = buffer.New()
	}
	if filename == "" {
		filename = "[Sem Nome]"
	} else {
		filename = filepath.Base(filename)
	}

	const defaultWidth = 80
	const defaultHeight = 24

	return &Editor{
		buf:      buf,
		viewport: view.NewViewport(defaultHeight - 1),
		filename: filename,
		width:    defaultWidth,
		height:   defaultHeight,
	}
}

func (e *Editor) Cursor() buffer.Pos {
	return e.cursor
}

func (e *Editor) HandleEvent(ev input.Event) {
	switch ev.Kind {
	case input.EventResize:
		e.width = ev.Width
		e.height = ev.Height
		vpHeight := e.height - 1
		if vpHeight < 1 {
			vpHeight = 1
		}
		e.viewport.Resize(vpHeight, e.buf.LineCount())
		e.viewport.EnsureVisible(e.cursor.Line, e.buf.LineCount())

	case input.EventKey:
		e.handleKey(ev.Key)
		e.viewport.EnsureVisible(e.cursor.Line, e.buf.LineCount())
	}
}

func (e *Editor) handleKey(k input.Key) {
	plain := func(r rune) bool {
		return k.Code == input.KeyRune && k.Rune == r && k.Mod == 0
	}

	switch {
	case k.Code == input.KeyDown || plain('j'):
		e.moveVertical(1)
	case k.Code == input.KeyUp || plain('k'):
		e.moveVertical(-1)
	case k.Code == input.KeyLeft || plain('h'):
		e.moveHorizontal(-1)
	case k.Code == input.KeyRight || plain('l'):
		e.moveHorizontal(1)
	case k.Code == input.KeyPageDown:
		e.moveVertical(e.viewport.Height)
	case k.Code == input.KeyPageUp:
		e.moveVertical(-e.viewport.Height)
	case plain('g'):
		e.moveToLine(0)
	case plain('G'):
		e.moveToLine(e.buf.LineCount() - 1)
	case k.Code == input.KeyHome || plain('0'):
		e.cursor.Col = 0
		e.wantCol = 0
	case k.Code == input.KeyEnd || plain('$'):
		e.cursor.Col = e.maxCol(e.cursor.Line)
		e.wantCol = math.MaxInt
	}
}

func (e *Editor) moveVertical(delta int) {
	e.moveToLine(e.cursor.Line + delta)
}

func (e *Editor) moveToLine(line int) {
	last := e.buf.LineCount() - 1
	if line < 0 {
		line = 0
	}
	if line > last {
		line = last
	}
	e.cursor.Line = line
	e.cursor.Col = min(e.wantCol, e.maxCol(line))
}

func (e *Editor) moveHorizontal(delta int) {
	col := e.cursor.Col + delta
	if col < 0 {
		col = 0
	}
	if maxCol := e.maxCol(e.cursor.Line); col > maxCol {
		col = maxCol
	}
	e.cursor.Col = col
	e.wantCol = col
}

func (e *Editor) maxCol(line int) int {
	n := e.buf.LineLen(line)
	if n == 0 {
		return 0
	}
	return n - 1
}

func (e *Editor) Frame() render.Frame {
	total := e.buf.LineCount()
	start, end := e.viewport.VisibleRange(total)

	vpHeight := e.viewport.Height
	lines := make([]render.StyledLine, 0, vpHeight)

	for i := start; i < end; i++ {
		formatted := view.FormatLine(e.buf.Line(i), e.width, view.DefaultTabStop)
		lines = append(lines, render.StyledLine{
			{Text: formatted, Style: render.Style{}},
		})
	}

	tildeStyle := render.Style{Fg: render.ColorBrightBlack}
	for len(lines) < vpHeight {
		lines = append(lines, render.StyledLine{
			{Text: "~", Style: tildeStyle},
		})
	}

	status := fmt.Sprintf(" %s   Ln %d/%d, Col %d", e.filename, e.cursor.Line+1, total, e.cursor.Col+1)

	return render.Frame{
		Lines:  lines,
		Cursor: e.screenCursor(start),
		Status: view.FormatLine([]rune(status), e.width, view.DefaultTabStop),
	}
}

func (e *Editor) screenCursor(top int) render.ScreenPos {
	col := view.ScreenCol(e.buf.Line(e.cursor.Line), e.cursor.Col, view.DefaultTabStop)
	if e.width > 0 && col > e.width-1 {
		col = e.width - 1
	}
	return render.ScreenPos{Row: e.cursor.Line - top, Col: col}
}
