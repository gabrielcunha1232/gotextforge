package engine

import "github.com/gabrielcunha1232/gotextforge/internal/core/buffer"

type Range struct {
	Start    buffer.Pos
	End      buffer.Pos
	Linewise bool
}

func TextObject(t Text, at buffer.Pos, m Motion) (Range, bool) {
	switch m.Object {
	case ObjectWord, ObjectBigWord:
		return wordObject(t.Line(at.Line), at, m.Inner, m.Object == ObjectBigWord)
	case ObjectParagraph:
		return paragraphObject(t, at.Line, m.Inner), true
	case ObjectDoubleQuote:
		return quoteObject(t.Line(at.Line), at, '"', m.Inner)
	case ObjectSingleQuote:
		return quoteObject(t.Line(at.Line), at, '\'', m.Inner)
	case ObjectBacktick:
		return quoteObject(t.Line(at.Line), at, '`', m.Inner)
	case ObjectParen:
		return bracketObject(t, at, '(', ')', m.Inner)
	case ObjectBracket:
		return bracketObject(t, at, '[', ']', m.Inner)
	case ObjectBrace:
		return bracketObject(t, at, '{', '}', m.Inner)
	case ObjectAngle:
		return bracketObject(t, at, '<', '>', m.Inner)
	}
	return Range{}, false
}

func wordObject(line []rune, at buffer.Pos, inner, big bool) (Range, bool) {
	if len(line) == 0 {
		return Range{}, false
	}
	col := min(at.Col, len(line)-1)
	cls := charClass(line[col], big)

	start, end := col, col+1
	for start > 0 && charClass(line[start-1], big) == cls {
		start--
	}
	for end < len(line) && charClass(line[end], big) == cls {
		end++
	}

	if !inner {
		if cls == 0 {
			if end < len(line) {
				next := charClass(line[end], big)
				for end < len(line) && charClass(line[end], big) == next {
					end++
				}
			}
		} else {
			trail := end
			for trail < len(line) && charClass(line[trail], big) == 0 {
				trail++
			}
			if trail > end {
				end = trail
			} else {
				for start > 0 && charClass(line[start-1], big) == 0 {
					start--
				}
			}
		}
	}

	return Range{Start: buffer.Pos{Line: at.Line, Col: start}, End: buffer.Pos{Line: at.Line, Col: end}}, true
}

func paragraphObject(t Text, line int, inner bool) Range {
	last := t.LineCount() - 1
	blank := func(l int) bool { return t.LineLen(l) == 0 }
	kind := blank(line)

	s, e := line, line
	for s > 0 && blank(s-1) == kind {
		s--
	}
	for e < last && blank(e+1) == kind {
		e++
	}

	if !inner {
		if e < last {
			next := blank(e + 1)
			for e < last && blank(e+1) == next {
				e++
			}
		} else if !kind {
			for s > 0 && blank(s-1) {
				s--
			}
		}
	}

	return Range{Start: buffer.Pos{Line: s}, End: buffer.Pos{Line: e, Col: t.LineLen(e)}, Linewise: true}
}

func quoteObject(line []rune, at buffer.Pos, q rune, inner bool) (Range, bool) {
	var quotes []int
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == q {
			quotes = append(quotes, i)
		}
	}

	open, close := -1, -1
	for i := 0; i+1 < len(quotes); i += 2 {
		if at.Col <= quotes[i+1] {
			open, close = quotes[i], quotes[i+1]
			break
		}
	}
	if open < 0 {
		return Range{}, false
	}

	if inner {
		return Range{Start: buffer.Pos{Line: at.Line, Col: open + 1}, End: buffer.Pos{Line: at.Line, Col: close}}, true
	}
	start, end := open, close+1
	trail := end
	for trail < len(line) && (line[trail] == ' ' || line[trail] == '\t') {
		trail++
	}
	if trail > end {
		end = trail
	} else {
		for start > 0 && (line[start-1] == ' ' || line[start-1] == '\t') {
			start--
		}
	}
	return Range{Start: buffer.Pos{Line: at.Line, Col: start}, End: buffer.Pos{Line: at.Line, Col: end}}, true
}

func findOpen(t Text, at buffer.Pos, open, close rune) (buffer.Pos, bool) {
	w := newWalker(t, at)
	depth := 0
	if w.char() == close {
		if !w.prev() {
			return buffer.Pos{}, false
		}
	}
	for {
		switch w.char() {
		case close:
			depth++
		case open:
			if depth == 0 {
				return w.pos(), true
			}
			depth--
		}
		if !w.prev() {
			return buffer.Pos{}, false
		}
	}
}

func findClose(t Text, openPos buffer.Pos, open, close rune) (buffer.Pos, bool) {
	w := newWalker(t, openPos)
	depth := 0
	for {
		if !w.next() {
			return buffer.Pos{}, false
		}
		switch w.char() {
		case open:
			depth++
		case close:
			if depth == 0 {
				return w.pos(), true
			}
			depth--
		}
	}
}

func bracketObject(t Text, at buffer.Pos, open, close rune, inner bool) (Range, bool) {
	openPos, ok := findOpen(t, at, open, close)
	if !ok {
		return Range{}, false
	}
	closePos, ok := findClose(t, openPos, open, close)
	if !ok {
		return Range{}, false
	}

	if !inner {
		return Range{Start: openPos, End: buffer.Pos{Line: closePos.Line, Col: closePos.Col + 1}}, true
	}

	openAtEOL := openPos.Col == t.LineLen(openPos.Line)-1
	closeLine := t.Line(closePos.Line)
	closeAfterBlank := true
	for _, r := range closeLine[:closePos.Col] {
		if r != ' ' && r != '\t' {
			closeAfterBlank = false
			break
		}
	}
	if openAtEOL && closeAfterBlank && closePos.Line > openPos.Line+1 {
		first, lastLine := openPos.Line+1, closePos.Line-1
		return Range{Start: buffer.Pos{Line: first}, End: buffer.Pos{Line: lastLine, Col: t.LineLen(lastLine)}, Linewise: true}, true
	}

	return Range{Start: buffer.Pos{Line: openPos.Line, Col: openPos.Col + 1}, End: closePos}, true
}

func FindOpen(t Text, at buffer.Pos, open, close rune) (buffer.Pos, bool) {
	return findOpen(t, at, open, close)
}
