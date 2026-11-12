package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type searchState struct {
	query     engine.Query
	forward   bool
	highlight bool
	promptFwd bool
}

func (e *Editor) promptLine() []rune {
	if e.mode == ModePrompt && e.prompt != nil {
		return append([]rune(e.prompt.label), e.prompt.text...)
	}
	prompt := ':'
	if e.mode == ModeSearch {
		prompt = '/'
		if !e.search.promptFwd {
			prompt = '?'
		}
	}
	return append([]rune{prompt}, e.cmdline...)
}

func (e *Editor) enterSearch(forward bool) {
	e.mode = ModeSearch
	e.search.promptFwd = forward
	e.cmdline = e.cmdline[:0]
}

func (e *Editor) executeSearch(pattern string) {
	if pattern != "" {
		e.search.query = engine.Query{Pattern: []rune(pattern)}
	}
	e.search.forward = e.search.promptFwd
	e.runSearch(e.search.forward, 1)
}

func (e *Editor) repeatSearch(same bool, count int) {
	forward := e.search.forward
	if !same {
		forward = !forward
	}
	e.runSearch(forward, count)
}

func (e *Editor) searchWord(forward bool, count int) {
	word, start, ok := engine.WordUnderCursor(e.buf.Line(e.cursor.Line), e.cursor.Col)
	if !ok {
		e.message = "E348: nenhuma palavra sob o cursor"
		return
	}
	e.search.query = engine.Query{Pattern: append([]rune(nil), word...), WholeWord: true}
	e.search.forward = forward
	e.cursor.Col = start
	e.runSearch(forward, count)
}

func (e *Editor) runSearch(forward bool, count int) {
	q := e.search.query
	if len(q.Pattern) == 0 {
		e.message = "E35: nenhum padrão de busca anterior"
		return
	}
	e.search.highlight = true
	pos, wrapped, ok := engine.Search(e.buf, e.cursor, q, forward, count)
	if !ok {
		e.message = "E486: padrão não encontrado: " + string(q.Pattern)
		return
	}
	e.cursor = pos
	e.setCol(pos.Col)
	switch {
	case wrapped && forward:
		e.message = "a busca chegou ao FIM e continuou do INÍCIO"
	case wrapped:
		e.message = "a busca chegou ao INÍCIO e continuou do FIM"
	default:
		prompt := "/"
		if !forward {
			prompt = "?"
		}
		e.message = prompt + string(q.Pattern)
	}
}

func (e *Editor) activeQuery() (engine.Query, bool) {
	if e.mode == ModeSearch {
		if len(e.cmdline) == 0 {
			return engine.Query{}, false
		}
		return engine.Query{Pattern: e.cmdline}, true
	}
	if e.search.highlight && len(e.search.query.Pattern) > 0 {
		return e.search.query, true
	}
	return engine.Query{}, false
}

func (e *Editor) applySearchHighlight(line int, runes []rune, styles []render.Style) {
	q, ok := e.activeQuery()
	if !ok {
		return
	}
	n := len(q.Pattern)
	for _, start := range q.Matches(runes) {
		style := styleSearch
		if e.mode != ModeSearch && line == e.cursor.Line && start == e.cursor.Col {
			style = styleSearchCurrent
		}
		for j := start; j < start+n && j < len(styles); j++ {
			styles[j].Fg = style.Fg
			styles[j].Bg = style.Bg
			styles[j].Bold = style.Bold
		}
	}
}
