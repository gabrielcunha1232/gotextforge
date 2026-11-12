package editor

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/storage"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
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
	buf       *buffer.Buffer
	viewport  view.Viewport
	cursor    buffer.Pos
	wantCol   int
	mode      Mode
	path      string
	filename  string
	save      SaveFunc
	cmdline   []rune
	message   string
	quit      bool
	highlight bool
	syntax    syntax.Cache
	parser    engine.Parser
	reg       register
	lsp       lspState
	hover     *hoverPopup
	comp      *completion
	width     int
	height    int
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
		buf:       buf,
		viewport:  view.NewViewport(defaultHeight - 1),
		path:      path,
		filename:  filename,
		save:      storage.Save,
		highlight: strings.HasSuffix(path, ".go"),
		width:     defaultWidth,
		height:    defaultHeight,
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
		if e.hover != nil {
			e.hover = nil
			if ev.Key.Code == input.KeyEsc {
				return
			}
		}
		switch e.mode {
		case ModeInsert:
			e.handleInsertKey(ev.Key)
			if e.mode == ModeInsert {
				e.afterInsertKey(ev.Key)
			}
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
	e.syncLanguageServer()
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
