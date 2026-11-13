package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
)

type LanguageClient interface {
	Open(path, text string, version int)
	Change(path, text string, version int)
	Save(path string)
	CloseDocument(path string)
	Hover(id int, path string, line []rune, pos buffer.Pos)
	Completion(id int, path string, line []rune, pos buffer.Pos)
	Events() <-chan lsp.Event
	Close()
}

type lspState struct {
	client   LanguageClient
	starter  func(path string) LanguageClient
	ready    bool
	failed   bool
	nextReq  int
	hoverReq int
	compReq  int
}

func (e *Editor) SetLanguageClient(c LanguageClient) {
	e.lsp = lspState{client: c}
	for _, d := range e.docs {
		d.opened = false
	}
	if c != nil {
		e.forward(c)
	}
	e.ensureLanguageServer(e.document)
}

func (e *Editor) forward(c LanguageClient) {
	events := c.Events()
	if events == nil {
		return
	}
	go func() {
		for ev := range events {
			e.async <- ev
		}
	}()
}

func (e *Editor) SetLanguageServer(start func(path string) LanguageClient) {
	e.lsp.starter = start
	e.ensureLanguageServer(e.document)
}

func (e *Editor) ensureLanguageServer(d *document) {
	if !d.highlight || d.path == "" || d.opened {
		return
	}
	if e.lsp.client == nil {
		if e.lsp.starter == nil {
			return
		}
		e.lsp.client = e.lsp.starter(d.path)
		e.forward(e.lsp.client)
	}
	d.synced = d.buf.Version()
	d.opened = true
	e.lsp.client.Open(d.path, d.buf.PlainText(), d.synced)
}

func (e *Editor) Async() <-chan any {
	return e.async
}

func (e *Editor) Close() {
	if e.term.job != nil {
		e.term.job.Cancel()
	}
	e.stopDockerRefresh()
	if e.lsp.client != nil {
		e.lsp.client.Close()
	}
}

func (e *Editor) HandleAsync(ev any) {
	switch v := ev.(type) {
	case lsp.StatusEvent:
		if v.Err != nil {
			e.lsp.failed = true
			e.message = "gopls: " + v.Err.Error()
			return
		}
		e.lsp.ready = v.Ready
	case lsp.DiagnosticsEvent:
		for _, d := range e.docs {
			if d.path != "" && lsp.SameURI(lsp.PathToURI(d.path), v.URI) {
				d.diags = v.Items
			}
		}
	case lsp.HoverEvent:
		if v.ID != e.lsp.hoverReq || e.mode != ModeNormal {
			return
		}
		switch {
		case v.Err != nil:
			e.message = "erro no hover: " + v.Err.Error()
		case v.Text == "":
			e.message = "nada a mostrar aqui"
		default:
			e.hover = newHoverPopup(v.Text)
		}
	case terminalTick:
		e.receiveTerminal(v)
	case dockerSnapshot:
		e.receiveDocker(v)
	case lsp.CompletionEvent:
		if v.ID != e.lsp.compReq || e.mode != ModeInsert || v.Err != nil {
			return
		}
		e.receiveCompletion(v.Items)
	}
}

func (e *Editor) syncLanguageServer() {
	if e.lsp.client == nil || !e.opened {
		return
	}
	if v := e.buf.Version(); v != e.synced {
		e.synced = v
		e.lsp.client.Change(e.path, e.buf.PlainText(), v)
	}
}

func (e *Editor) requestHover() {
	if e.lsp.client == nil || !e.opened {
		e.message = "sem servidor de linguagem (abra um arquivo .go com o gopls instalado)"
		return
	}
	e.syncLanguageServer()
	e.lsp.nextReq++
	e.lsp.hoverReq = e.lsp.nextReq
	e.lsp.client.Hover(e.lsp.hoverReq, e.path, e.buf.Line(e.cursor.Line), e.cursor)
}

func (e *Editor) requestCompletion() {
	if e.lsp.client == nil || !e.lsp.ready || !e.opened {
		return
	}
	e.syncLanguageServer()
	e.lsp.nextReq++
	e.lsp.compReq = e.lsp.nextReq
	e.lsp.client.Completion(e.lsp.compReq, e.path, e.buf.Line(e.cursor.Line), e.cursor)
}
