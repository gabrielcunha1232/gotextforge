package editor

import (
	"sort"
	"strings"
	"unicode"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

const completionMaxRows = 10

type completion struct {
	all   []lsp.CompletionItem
	items []lsp.CompletionItem
	sel   int
	start buffer.Pos
}

func isIdentRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (e *Editor) prefixStart() buffer.Pos {
	line := e.buf.Line(e.cursor.Line)
	col := min(e.cursor.Col, len(line))
	for col > 0 && isIdentRune(line[col-1]) {
		col--
	}
	return buffer.Pos{Line: e.cursor.Line, Col: col}
}

func (e *Editor) receiveCompletion(items []lsp.CompletionItem) {
	sorted := append([]lsp.CompletionItem(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SortText != sorted[j].SortText {
			return sorted[i].SortText < sorted[j].SortText
		}
		return sorted[i].Label < sorted[j].Label
	})
	e.comp = &completion{all: sorted, start: e.prefixStart()}
	e.filterCompletion()
}

func (e *Editor) filterCompletion() {
	if e.comp == nil {
		return
	}
	start := e.prefixStart()
	if start != e.comp.start {
		e.comp = nil
		return
	}
	line := e.buf.Line(e.cursor.Line)
	prefix := strings.ToLower(string(line[start.Col:min(e.cursor.Col, len(line))]))

	e.comp.items = e.comp.items[:0]
	for _, it := range e.comp.all {
		if strings.HasPrefix(strings.ToLower(it.Filter()), prefix) {
			e.comp.items = append(e.comp.items, it)
		}
	}
	if len(e.comp.items) == 0 {
		e.comp = nil
		return
	}
	e.comp.sel = min(e.comp.sel, len(e.comp.items)-1)
}

func (e *Editor) handleCompletionKey(k input.Key) bool {
	if e.comp == nil {
		if isCtrl(k, ' ') || isCtrl(k, 'n') {
			e.requestCompletion()
			return true
		}
		return false
	}

	n := len(e.comp.items)
	switch {
	case isCtrl(k, 'n') || k.Code == input.KeyDown:
		e.comp.sel = (e.comp.sel + 1) % n
	case isCtrl(k, 'p') || k.Code == input.KeyUp:
		e.comp.sel = (e.comp.sel - 1 + n) % n
	case k.Code == input.KeyEnter || k.Code == input.KeyTab || isCtrl(k, 'y'):
		e.acceptCompletion()
	case isCtrl(k, 'e'):
		e.comp = nil
	default:
		if k.Code != input.KeyRune && k.Code != input.KeyBackspace {
			e.comp = nil
		}
		return false
	}
	return true
}

func (e *Editor) acceptCompletion() {
	item := e.comp.items[e.comp.sel]
	start := e.comp.start
	e.comp = nil
	e.buf.Delete(start, e.cursor)
	e.cursor = e.buf.Insert(start, item.Text())
	e.wantCol = e.cursor.Col
}

func (e *Editor) afterInsertKey(k input.Key) {
	switch {
	case k.Code == input.KeyRune && k.Mod == 0 && k.Rune == '.':
		e.comp = nil
		e.requestCompletion()
	case k.Code == input.KeyRune && k.Mod == 0 && isIdentRune(k.Rune):
		if e.comp != nil {
			e.filterCompletion()
			return
		}
		if e.cursor.Col-e.prefixStart().Col == 1 {
			e.requestCompletion()
		}
	case k.Code == input.KeyBackspace:
		e.filterCompletion()
	case k.Code == input.KeyRune && k.Mod == 0:
		e.comp = nil
	}
}

var kindTags = map[int]string{
	1: "text", 2: "meth", 3: "func", 4: "ctor", 5: "field", 6: "var", 7: "class", 8: "iface",
	9: "pkg", 10: "prop", 11: "unit", 12: "value", 13: "enum", 14: "kw", 15: "snip", 16: "color",
	17: "file", 18: "ref", 19: "dir", 20: "enum", 21: "const", 22: "struct", 23: "event", 24: "op", 25: "type",
}

func kindColor(kind int) render.Color {
	switch kind {
	case 2, 3, 4:
		return colBlue
	case 5, 10:
		return colGreen
	case 6, 12:
		return colMagenta
	case 7, 8, 22, 25:
		return colBlue1
	case 9:
		return colOrange
	case 14:
		return colMagenta
	case 21, 13, 20:
		return colOrange
	default:
		return colFgDark
	}
}

func (e *Editor) completionBox() []render.StyledLine {
	c := e.comp
	first := 0
	if c.sel >= completionMaxRows {
		first = c.sel - completionMaxRows + 1
	}
	last := min(first+completionMaxRows, len(c.items))

	labelW, detailW := 0, 0
	for _, it := range c.items[first:last] {
		labelW = max(labelW, len([]rune(sanitize(it.Label))))
		detailW = max(detailW, len([]rune(sanitize(it.Detail))))
	}
	labelW = min(labelW, 40)
	detailW = min(detailW, 30)

	var box []render.StyledLine
	for i := first; i < last; i++ {
		it := c.items[i]
		bg := colBgFloat
		if i == c.sel {
			bg = colBgSel
		}
		tag := kindTags[it.Kind]
		row := render.StyledLine{
			{Text: " " + padRight(tag, 6), Style: render.Style{Fg: kindColor(it.Kind), Bg: bg}},
			{Text: padRight(truncateText(sanitize(it.Label), labelW), labelW), Style: render.Style{Fg: colFg, Bg: bg, Bold: i == c.sel}},
		}
		if detailW > 0 {
			row = append(row, render.Span{Text: "  " + padRight(truncateText(sanitize(it.Detail), detailW), detailW), Style: render.Style{Fg: colComment, Bg: bg}})
		}
		row = append(row, render.Span{Text: " ", Style: render.Style{Bg: bg}})
		box = append(box, row)
	}
	return box
}

func truncateText(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}
