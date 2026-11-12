package engine

import (
	"unicode"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

type Text interface {
	LineCount() int
	LineLen(i int) int
	Line(i int) []rune
}

type Target struct {
	Pos       buffer.Pos
	Linewise  bool
	Inclusive bool
	Vertical  bool
}

type MotionContext struct {
	Count    int
	HasCount bool
	Page     int
	Operator bool
}

func ApplyMotion(t Text, from buffer.Pos, m Motion, ctx MotionContext) (Target, bool) {
	n := max(ctx.Count, 1)
	last := t.LineCount() - 1
	lineLen := t.LineLen(from.Line)

	switch m.Kind {
	case MotionLeft:
		if from.Col == 0 {
			return Target{}, false
		}
		return Target{Pos: buffer.Pos{Line: from.Line, Col: max(from.Col-n, 0)}}, true

	case MotionRight:
		limit := lineLen - 1
		if ctx.Operator {
			limit = lineLen
		}
		if from.Col >= limit {
			return Target{}, false
		}
		return Target{Pos: buffer.Pos{Line: from.Line, Col: min(from.Col+n, limit)}}, true

	case MotionDown, MotionUp:
		delta := n
		if m.Kind == MotionUp {
			delta = -n
		}
		line := clamp(from.Line+delta, 0, last)
		if line == from.Line {
			return Target{}, false
		}
		return Target{Pos: buffer.Pos{Line: line}, Linewise: true, Vertical: true}, true

	case MotionPageDown, MotionPageUp:
		delta := max(ctx.Page, 1) * n
		if m.Kind == MotionPageUp {
			delta = -delta
		}
		return Target{Pos: buffer.Pos{Line: clamp(from.Line+delta, 0, last)}, Linewise: true, Vertical: true}, true

	case MotionFileStart, MotionFileEnd:
		line := last
		if m.Kind == MotionFileStart {
			line = 0
		}
		if ctx.HasCount {
			line = clamp(ctx.Count-1, 0, last)
		}
		return Target{Pos: buffer.Pos{Line: line, Col: firstNonBlank(t.Line(line))}, Linewise: true}, true

	case MotionLineStart:
		return Target{Pos: buffer.Pos{Line: from.Line}}, true

	case MotionFirstNonBlank:
		return Target{Pos: buffer.Pos{Line: from.Line, Col: firstNonBlank(t.Line(from.Line))}}, true

	case MotionLineEnd:
		line := clamp(from.Line+n-1, 0, last)
		return Target{Pos: buffer.Pos{Line: line, Col: max(t.LineLen(line)-1, 0)}, Inclusive: true}, true

	case MotionFindForward, MotionTillForward, MotionFindBackward, MotionTillBackward:
		return findInLine(t.Line(from.Line), from, m, n)

	case MotionWordStart, MotionBigWordStart:
		return Target{Pos: wordStart(t, from, n, m.Kind == MotionBigWordStart)}, true

	case MotionWordEnd, MotionBigWordEnd:
		return Target{Pos: wordEnd(t, from, n, m.Kind == MotionBigWordEnd), Inclusive: true}, true

	case MotionWordBack, MotionBigWordBack:
		p := wordBack(t, from, n, m.Kind == MotionBigWordBack)
		if p == from {
			return Target{}, false
		}
		return Target{Pos: p}, true

	case MotionParagraphForward, MotionParagraphBackward:
		return Target{Pos: paragraphMove(t, from, n, m.Kind == MotionParagraphForward)}, true
	}
	return Target{}, false
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

func firstNonBlank(line []rune) int {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return i
		}
	}
	return max(len(line)-1, 0)
}

func findInLine(line []rune, from buffer.Pos, m Motion, n int) (Target, bool) {
	forward := m.Kind == MotionFindForward || m.Kind == MotionTillForward
	till := m.Kind == MotionTillForward || m.Kind == MotionTillBackward
	col := from.Col
	start := col
	if till {
		if forward {
			start = col + 1
		} else {
			start = col - 1
		}
	}

	found := -1
	i := start
	for count := 0; count < n; count++ {
		found = -1
		for {
			if forward {
				i++
			} else {
				i--
			}
			if i < 0 || i >= len(line) {
				return Target{}, false
			}
			if line[i] == m.Char {
				found = i
				break
			}
		}
	}

	if till {
		if forward {
			found--
		} else {
			found++
		}
	}
	if found == col && till {
		return Target{}, false
	}
	return Target{Pos: buffer.Pos{Line: from.Line, Col: found}, Inclusive: forward}, true
}

func charClass(r rune, big bool) int {
	switch {
	case r == ' ' || r == '\t' || r == '\n':
		return 0
	case big:
		return 1
	case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
		return 2
	default:
		return 1
	}
}

type walker struct {
	t       Text
	line    int
	col     int
	runes   []rune
	runesOf int
}

func newWalker(t Text, p buffer.Pos) *walker {
	w := &walker{t: t, line: p.Line, col: p.Col, runesOf: -1}
	w.load()
	if w.col > len(w.runes) {
		w.col = len(w.runes)
	}
	return w
}

func (w *walker) load() {
	if w.runesOf != w.line {
		w.runes = w.t.Line(w.line)
		w.runesOf = w.line
	}
}

func (w *walker) char() rune {
	if w.col >= len(w.runes) {
		return '\n'
	}
	return w.runes[w.col]
}

func (w *walker) emptyLine() bool {
	return len(w.runes) == 0
}

func (w *walker) next() bool {
	if w.col < len(w.runes) {
		w.col++
		return true
	}
	if w.line+1 >= w.t.LineCount() {
		return false
	}
	w.line++
	w.col = 0
	w.load()
	return true
}

func (w *walker) prev() bool {
	if w.col > 0 {
		w.col--
		return true
	}
	if w.line == 0 {
		return false
	}
	w.line--
	w.load()
	w.col = len(w.runes)
	return true
}

func (w *walker) pos() buffer.Pos {
	return buffer.Pos{Line: w.line, Col: w.col}
}

func wordStart(t Text, from buffer.Pos, n int, big bool) buffer.Pos {
	w := newWalker(t, from)
	for ; n > 0; n-- {
		startLine := w.line
		if cls := charClass(w.char(), big); cls != 0 {
			for charClass(w.char(), big) == cls {
				if !w.next() {
					return w.pos()
				}
			}
		}
		for charClass(w.char(), big) == 0 {
			if w.emptyLine() && w.line != startLine {
				break
			}
			if !w.next() {
				return w.pos()
			}
		}
	}
	return w.pos()
}

func wordEnd(t Text, from buffer.Pos, n int, big bool) buffer.Pos {
	w := newWalker(t, from)
	for ; n > 0; n-- {
		if !w.next() {
			break
		}
		for charClass(w.char(), big) == 0 {
			if !w.next() {
				return w.pos()
			}
		}
		cls := charClass(w.char(), big)
		for w.col+1 < len(w.runes) && charClass(w.runes[w.col+1], big) == cls {
			w.col++
		}
	}
	return w.pos()
}

func wordBack(t Text, from buffer.Pos, n int, big bool) buffer.Pos {
	w := newWalker(t, from)
	for ; n > 0; n-- {
		if !w.prev() {
			break
		}
		for charClass(w.char(), big) == 0 {
			if w.emptyLine() {
				break
			}
			if !w.prev() {
				return w.pos()
			}
		}
		if w.emptyLine() {
			continue
		}
		cls := charClass(w.char(), big)
		for w.col > 0 && charClass(w.runes[w.col-1], big) == cls {
			w.col--
		}
	}
	return w.pos()
}

func paragraphMove(t Text, from buffer.Pos, n int, forward bool) buffer.Pos {
	last := t.LineCount() - 1
	l := from.Line
	step := 1
	if !forward {
		step = -1
	}
	canStep := func(l int) bool { return l+step >= 0 && l+step <= last }
	blank := func(l int) bool { return t.LineLen(l) == 0 }

	for ; n > 0; n-- {
		for blank(l) && canStep(l) {
			l += step
		}
		for !blank(l) && canStep(l) {
			l += step
		}
	}

	switch {
	case forward && t.LineLen(l) != 0:
		return buffer.Pos{Line: l, Col: t.LineLen(l)}
	default:
		return buffer.Pos{Line: l}
	}
}

func ChangeWordTarget(t Text, from buffer.Pos, count int, big bool) (Target, bool) {
	line := t.Line(from.Line)
	if from.Col >= len(line) {
		return Target{}, false
	}
	cls := charClass(line[from.Col], big)
	if cls == 0 {
		return Target{}, false
	}
	end := from.Col
	for end+1 < len(line) && charClass(line[end+1], big) == cls {
		end++
	}
	p := buffer.Pos{Line: from.Line, Col: end}
	if count > 1 {
		p = wordEnd(t, p, count-1, big)
	}
	return Target{Pos: p, Inclusive: true}, true
}

func FirstNonBlank(line []rune) int {
	return firstNonBlank(line)
}
