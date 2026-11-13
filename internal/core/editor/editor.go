package editor

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
	"github.com/gabrielcunha1232/gotextforge/internal/core/storage"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

type Mode uint8

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeSearch
	ModePrompt
	ModeExplorer
	ModeFinder
	ModeTerminal
	ModeDocker
)

type SaveFunc func(path string, src io.WriterTo) error

type LoadFunc func(path string) (*buffer.Buffer, error)

type document struct {
	buf       *buffer.Buffer
	path      string
	filename  string
	highlight bool
	syntax    syntax.Cache
	viewport  view.Viewport
	cursor    buffer.Pos
	wantCol   int
	diags     []lsp.Diagnostic
	synced    int
	opened    bool
}

func newDocument(buf *buffer.Buffer, path string) *document {
	if buf == nil {
		buf = buffer.New()
	}
	filename := "[Sem Nome]"
	if path != "" {
		filename = filepath.Base(path)
	}
	return &document{
		buf:       buf,
		path:      path,
		filename:  filename,
		highlight: strings.HasSuffix(path, ".go"),
		viewport:  view.NewViewport(1),
	}
}

type Editor struct {
	*document
	docs       []*document
	mode       Mode
	save       SaveFunc
	load       LoadFunc
	cmdline    []rune
	message    string
	quit       bool
	parser     engine.Parser
	reg        register
	lsp        lspState
	hover      *hoverPopup
	comp       *completion
	search     searchState
	root       string
	term       terminal
	start      StartFunc
	dock       dockerState
	dockerExec func(dir string) docker.Exec
	async      chan any
	explorer   explorer
	finder     *finder
	prompt     *prompt
	width      int
	height     int
}

func New(buf *buffer.Buffer, path string) *Editor {
	doc := newDocument(buf, path)
	e := &Editor{
		document:   doc,
		docs:       []*document{doc},
		save:       storage.Save,
		load:       loadFile,
		async:      make(chan any, 256),
		start:      runner.Start,
		dockerExec: defaultDockerExec,
		width:      80,
		height:     24,
	}
	e.layout()
	return e
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
		e.layout()

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
		case ModePrompt:
			e.handlePromptKey(ev.Key)
		case ModeExplorer:
			e.handleExplorerKey(ev.Key)
		case ModeFinder:
			e.handleFinderKey(ev.Key)
		case ModeTerminal:
			e.handleTerminalKey(ev.Key)
		case ModeDocker:
			e.handleDockerKey(ev.Key)
		case ModeCommand, ModeSearch:
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
