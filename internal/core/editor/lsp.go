package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
)

type LanguageClient interface {
	Open(text string, version int)
	Change(text string, version int)
	Save()
	Hover(id int, line []rune, pos buffer.Pos)
	Completion(id int, line []rune, pos buffer.Pos)
	Events() <-chan lsp.Event
	Close()
}

type lspState struct {
	client   LanguageClient
	ready    bool
	failed   bool
	synced   int
	diags    []lsp.Diagnostic
	nextReq  int
	hoverReq int
	compReq  int
}

func (e *Editor) SetLanguageClient(c LanguageClient) {
	e.lsp = lspState{client: c}
	if c == nil {
		return
	}
	e.lsp.synced = e.buf.Version()
	c.Open(e.buf.PlainText(), e.lsp.synced)
}

func (e *Editor) Async() <-chan lsp.Event {
	if e.lsp.client == nil {
		return nil
	}
	return e.lsp.client.Events()
}

func (e *Editor) Close() {
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
		e.lsp.diags = v.Items
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
	case lsp.CompletionEvent:
		if v.ID != e.lsp.compReq || e.mode != ModeInsert || v.Err != nil {
			return
		}
		e.receiveCompletion(v.Items)
	}
}

func (e *Editor) syncLanguageServer() {
	if e.lsp.client == nil {
		return
	}
	if v := e.buf.Version(); v != e.lsp.synced {
		e.lsp.synced = v
		e.lsp.client.Change(e.buf.PlainText(), v)
	}
}

func (e *Editor) requestHover() {
	if e.lsp.client == nil {
		e.message = "sem servidor de linguagem (abra um arquivo .go com o gopls instalado)"
		return
	}
	e.syncLanguageServer()
	e.lsp.nextReq++
	e.lsp.hoverReq = e.lsp.nextReq
	e.lsp.client.Hover(e.lsp.hoverReq, e.buf.Line(e.cursor.Line), e.cursor)
}

func (e *Editor) requestCompletion() {
	if e.lsp.client == nil || !e.lsp.ready {
		return
	}
	e.syncLanguageServer()
	e.lsp.nextReq++
	e.lsp.compReq = e.lsp.nextReq
	e.lsp.client.Completion(e.lsp.compReq, e.buf.Line(e.cursor.Line), e.cursor)
}
