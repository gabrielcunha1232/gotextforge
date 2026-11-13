package editor

import (
	"strings"
	"unicode"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
)

var pairs = map[rune]rune{'(': ')', '[': ']', '{': '}'}

var closers = map[rune]rune{')': '(', ']': '[', '}': '{'}

var quotes = map[rune]bool{'"': true, '\'': true, '`': true}

func runeAt(line []rune, i int) rune {
	if i < 0 || i >= len(line) {
		return 0
	}
	return line[i]
}

func leadingWhitespace(line []rune) []rune {
	n := 0
	for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
		n++
	}
	return line[:n]
}

func onlyBlank(runes []rune) bool {
	for _, r := range runes {
		if r != ' ' && r != '\t' {
			return false
		}
	}
	return true
}

func canPairBefore(next rune) bool {
	return next == 0 || unicode.IsSpace(next) || closers[next] != 0 || next == ',' || next == ';' || next == ':'
}

func (e *Editor) indentUnit(line []rune) string {
	if e.highlight {
		return "\t"
	}
	if ws := leadingWhitespace(line); len(ws) > 0 && ws[0] == ' ' {
		return "    "
	}
	return "\t"
}

func (e *Editor) moveRight() {
	e.cursor.Col++
	e.wantCol = e.cursor.Col
}

func (e *Editor) typeRune(r rune) {
	line := e.buf.Line(e.cursor.Line)
	col := e.cursor.Col
	prev, next := runeAt(line, col-1), runeAt(line, col)

	switch {
	case (closers[r] != 0 || quotes[r]) && next == r:
		e.moveRight()
	case pairs[r] != 0 && canPairBefore(next):
		e.insertText(string(r) + string(pairs[r]))
		e.cursor.Col--
		e.wantCol = e.cursor.Col
	case quotes[r] && canPairBefore(next) && !isIdentRune(prev) && prev != r && prev != '\\':
		e.insertText(string(r) + string(r))
		e.cursor.Col--
		e.wantCol = e.cursor.Col
	case closers[r] != 0 && onlyBlank(line[:col]):
		e.dedentForCloser(r)
		e.insertText(string(r))
	default:
		e.insertText(string(r))
	}
}

func (e *Editor) dedentForCloser(r rune) {
	at := e.cursor
	if at.Col > 0 {
		at.Col--
	} else if at.Line > 0 {
		at = buffer.Pos{Line: at.Line - 1, Col: e.buf.LineLen(at.Line - 1)}
	}
	open, ok := engine.FindOpen(e.buf, at, closers[r], r)
	if !ok {
		return
	}
	indent := string(leadingWhitespace(e.buf.Line(open.Line)))
	start := buffer.Pos{Line: e.cursor.Line}
	e.buf.Delete(start, e.cursor)
	e.cursor = e.buf.Insert(start, indent)
	e.wantCol = e.cursor.Col
}

func (e *Editor) backspacePair() bool {
	line := e.buf.Line(e.cursor.Line)
	col := e.cursor.Col
	prev, next := runeAt(line, col-1), runeAt(line, col)
	if prev == 0 || next == 0 {
		return false
	}
	if pairs[prev] != next && !(quotes[prev] && prev == next) {
		return false
	}
	start := buffer.Pos{Line: e.cursor.Line, Col: col - 1}
	e.buf.Delete(start, buffer.Pos{Line: e.cursor.Line, Col: col + 1})
	e.cursor = start
	e.wantCol = start.Col
	return true
}

func (e *Editor) newline() {
	line := e.buf.Line(e.cursor.Line)
	col := min(e.cursor.Col, len(line))
	before := line[:col]
	indent := string(leadingWhitespace(before))
	unit := e.indentUnit(line)

	trimmed := strings.TrimRight(string(before), " \t")
	opener := rune(0)
	if t := []rune(trimmed); len(t) > 0 {
		opener = t[len(t)-1]
	}
	next := runeAt(line, col)

	if onlyBlank(before) && onlyBlank(line[col:]) && col > 0 {
		start := buffer.Pos{Line: e.cursor.Line}
		e.buf.Delete(start, buffer.Pos{Line: e.cursor.Line, Col: len(line)})
		e.cursor = start
	}

	switch {
	case pairs[opener] != 0 && next == pairs[opener]:
		inner := indent + unit
		e.insertText("\n" + inner + "\n" + indent)
		e.cursor = buffer.Pos{Line: e.cursor.Line - 1, Col: len([]rune(inner))}
		e.wantCol = e.cursor.Col
	case pairs[opener] != 0:
		e.insertText("\n" + indent + unit)
	default:
		e.insertText("\n" + indent)
	}
}

func (e *Editor) openLineIndent(below bool) string {
	line := e.buf.Line(e.cursor.Line)
	indent := string(leadingWhitespace(line))
	unit := e.indentUnit(line)
	t := []rune(strings.TrimSpace(string(line)))
	if len(t) == 0 {
		return indent
	}
	if below && pairs[t[len(t)-1]] != 0 {
		return indent + unit
	}
	if !below && closers[t[0]] != 0 {
		return indent + unit
	}
	return indent
}
