package editor

import (
	"fmt"
	"io"
	"math"
	"path/filepath"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/storage"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

type Mode uint8

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
)

type SaveFunc func(path string, src io.WriterTo) error

type Editor struct {
	buf      *buffer.Buffer
	viewport view.Viewport
	cursor   buffer.Pos
	wantCol  int
	mode     Mode
	path     string
	filename string
	save     SaveFunc
	cmdline  []rune
	message  string
	quit     bool
	width    int
	height   int
}

func New(buf *buffer.Buffer, path string) *Editor {
	if buf == nil {
		buf = buffer.New()
	}
	filename := "[Sem Nome]"
	if path != "" {
		filename = filepath.Base(path)
	}

	const defaultWidth = 80
	const defaultHeight = 24

	return &Editor{
		buf:      buf,
		viewport: view.NewViewport(defaultHeight - 1),
		path:     path,
		filename: filename,
		save:     storage.Save,
		width:    defaultWidth,
		height:   defaultHeight,
	}
}

func (e *Editor) ShouldQuit() bool {
	return e.quit
}

func (e *Editor) Cursor() buffer.Pos {
	return e.cursor
}

func (e *Editor) Mode() Mode {
	return e.mode
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

	case input.EventKey:
		e.message = ""
		switch e.mode {
		case ModeInsert:
			e.handleInsertKey(ev.Key)
		case ModeCommand:
			e.handleCommandKey(ev.Key)
		default:
			e.buf.BeginChange(e.cursor)
			e.handleNormalKey(ev.Key)
			if e.mode != ModeInsert {
				e.buf.EndChange()
			}
		}
	}
	e.viewport.EnsureVisible(e.cursor.Line, e.buf.LineCount())
}

func (e *Editor) handleNormalKey(k input.Key) {
	if e.handleMotion(k) {
		return
	}

	switch {
	case isPlain(k, ':'):
		e.enterCommand()
	case isCtrl(k, 'c'):
		e.requestQuit()
	case isPlain(k, 'u'):
		pos, ok := e.buf.Undo()
		e.afterHistory(pos, ok, "Já na alteração mais antiga")
	case isCtrl(k, 'r'):
		pos, ok := e.buf.Redo()
		e.afterHistory(pos, ok, "Já na alteração mais recente")
	case isPlain(k, 'i'):
		e.enterInsert(e.cursor)
	case isPlain(k, 'a'):
		e.enterInsert(buffer.Pos{Line: e.cursor.Line, Col: min(e.cursor.Col+1, e.buf.LineLen(e.cursor.Line))})
	case isPlain(k, 'I'):
		e.enterInsert(buffer.Pos{Line: e.cursor.Line, Col: 0})
	case isPlain(k, 'A'):
		e.enterInsert(buffer.Pos{Line: e.cursor.Line, Col: e.buf.LineLen(e.cursor.Line)})
	case isPlain(k, 'o'):
		end := e.buf.Insert(buffer.Pos{Line: e.cursor.Line, Col: e.buf.LineLen(e.cursor.Line)}, "\n")
		e.enterInsert(end)
	case isPlain(k, 'O'):
		e.buf.Insert(buffer.Pos{Line: e.cursor.Line, Col: 0}, "\n")
		e.enterInsert(buffer.Pos{Line: e.cursor.Line, Col: 0})
	case isPlain(k, 'x') || (k.Code == input.KeyDelete && k.Mod == 0):
		if e.buf.LineLen(e.cursor.Line) > 0 {
			e.buf.Delete(e.cursor, buffer.Pos{Line: e.cursor.Line, Col: e.cursor.Col + 1})
			e.setCol(min(e.cursor.Col, e.maxCol(e.cursor.Line)))
		}
	}
}

func (e *Editor) handleMotion(k input.Key) bool {
	switch {
	case k.Code == input.KeyDown || (e.mode == ModeNormal && isPlain(k, 'j')):
		e.moveVertical(1)
	case k.Code == input.KeyUp || (e.mode == ModeNormal && isPlain(k, 'k')):
		e.moveVertical(-1)
	case k.Code == input.KeyLeft || (e.mode == ModeNormal && isPlain(k, 'h')):
		e.moveHorizontal(-1)
	case k.Code == input.KeyRight || (e.mode == ModeNormal && isPlain(k, 'l')):
		e.moveHorizontal(1)
	case k.Code == input.KeyPageDown:
		e.moveVertical(e.viewport.Height)
	case k.Code == input.KeyPageUp:
		e.moveVertical(-e.viewport.Height)
	case e.mode == ModeNormal && isPlain(k, 'g'):
		e.moveToLine(0)
	case e.mode == ModeNormal && isPlain(k, 'G'):
		e.moveToLine(e.buf.LineCount() - 1)
	case k.Code == input.KeyHome || (e.mode == ModeNormal && isPlain(k, '0')):
		e.setCol(0)
	case k.Code == input.KeyEnd || (e.mode == ModeNormal && isPlain(k, '$')):
		e.cursor.Col = e.maxCol(e.cursor.Line)
		e.wantCol = math.MaxInt
	default:
		return false
	}
	return true
}

func (e *Editor) afterHistory(pos buffer.Pos, ok bool, emptyMsg string) {
	if !ok {
		e.message = emptyMsg
		return
	}
	e.cursor.Line = pos.Line
	e.setCol(min(pos.Col, e.maxCol(pos.Line)))
}

func isPlain(k input.Key, r rune) bool {
	return k.Code == input.KeyRune && k.Rune == r && k.Mod == 0
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
	e.setCol(min(col, e.maxCol(e.cursor.Line)))
}

func (e *Editor) setCol(col int) {
	e.cursor.Col = col
	e.wantCol = col
}

func (e *Editor) maxCol(line int) int {
	n := e.buf.LineLen(line)
	if e.mode == ModeInsert {
		return n
	}
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

	frame := render.Frame{
		Lines:       lines,
		Cursor:      e.screenCursor(start),
		CursorShape: render.CursorBlock,
		Status:      view.FormatLine([]rune(e.status()), e.width, view.DefaultTabStop),
	}

	switch e.mode {
	case ModeInsert:
		frame.CursorShape = render.CursorBar
	case ModeCommand:
		cmd := append([]rune{':'}, e.cmdline...)
		frame.Status = view.FormatLine(cmd, e.width, view.DefaultTabStop)
		frame.CursorShape = render.CursorBar
		frame.Cursor = render.ScreenPos{
			Row: vpHeight,
			Col: min(view.ScreenCol(cmd, len(cmd), view.DefaultTabStop), max(e.width-1, 0)),
		}
	}

	return frame
}

func (e *Editor) status() string {
	if e.message != "" {
		return e.message
	}
	mode := ""
	if e.mode == ModeInsert {
		mode = " -- INSERT --"
	}
	modified := ""
	if e.buf.Modified() {
		modified = " [+]"
	}
	return fmt.Sprintf("%s %s%s   Ln %d/%d, Col %d",
		mode, e.filename, modified, e.cursor.Line+1, e.buf.LineCount(), e.cursor.Col+1)
}

func (e *Editor) screenCursor(top int) render.ScreenPos {
	col := view.ScreenCol(e.buf.Line(e.cursor.Line), e.cursor.Col, view.DefaultTabStop)
	if e.width > 0 && col > e.width-1 {
		col = e.width - 1
	}
	return render.ScreenPos{Row: e.cursor.Line - top, Col: col}
}
