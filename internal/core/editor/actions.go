package editor

import (
	"fmt"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func (e *Editor) escape() {
	e.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyEsc}))
}

func (e *Editor) SwitchTab(i int) {
	e.escape()
	if i < 0 || i >= len(e.docs) {
		return
	}
	e.switchTo(i)
	e.viewport.EnsureVisible(e.cursor.Line, e.buf.LineCount())
	e.syncLanguageServer()
}

func (e *Editor) CloseTab(i int) {
	e.escape()
	if i < 0 || i >= len(e.docs) {
		return
	}
	e.switchTo(i)
	if !e.buf.Modified() {
		e.closeTab()
		return
	}
	label := fmt.Sprintf("%s tem alterações. Salvar? (s)im, (n)ão, (c)ancelar: ", e.filename)
	e.openPrompt(label, "", func(answer string) {
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "s", "sim":
			if e.write() {
				e.closeTab()
			}
		case "n", "não", "nao":
			e.closeTab()
		default:
			e.message = "fechar cancelado"
		}
	})
}

func (e *Editor) closeTab() {
	if len(e.docs) > 1 {
		e.closeCurrent(true)
		return
	}
	if e.lsp.client != nil && e.opened {
		e.lsp.client.CloseDocument(e.path)
	}
	doc := newDocument(nil, "")
	e.docs = []*document{doc}
	e.document = doc
	e.layout()
}
