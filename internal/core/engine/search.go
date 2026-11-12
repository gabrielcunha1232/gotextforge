package engine

import (
	"unicode"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

type Query struct {
	Pattern   []rune
	WholeWord bool
}

func (q Query) ignoreCase() bool {
	for _, r := range q.Pattern {
		if unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func (q Query) Matches(line []rune) []int {
	n := len(q.Pattern)
	if n == 0 || n > len(line) {
		return nil
	}
	fold := q.ignoreCase()
	var out []int
	for i := 0; i+n <= len(line); i++ {
		if !q.matchAt(line, i, fold) {
			continue
		}
		if q.WholeWord && !isWordBoundary(line, i, i+n) {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (q Query) matchAt(line []rune, at int, fold bool) bool {
	for j, p := range q.Pattern {
		r := line[at+j]
		if r == p {
			continue
		}
		if !fold || unicode.ToLower(r) != unicode.ToLower(p) {
			return false
		}
	}
	return true
}

func isWordBoundary(line []rune, start, end int) bool {
	before := start == 0 || charClass(line[start-1], false) != 2
	after := end >= len(line) || charClass(line[end], false) != 2
	return before && after
}

func Search(t Text, from buffer.Pos, q Query, forward bool, count int) (buffer.Pos, bool, bool) {
	if len(q.Pattern) == 0 {
		return from, false, false
	}
	pos := from
	wrapped := false
	for n := max(count, 1); n > 0; n-- {
		next, w, ok := searchOnce(t, pos, q, forward)
		if !ok {
			return from, false, false
		}
		pos = next
		wrapped = wrapped || w
	}
	return pos, wrapped, true
}

func searchOnce(t Text, from buffer.Pos, q Query, forward bool) (buffer.Pos, bool, bool) {
	total := t.LineCount()
	for step := 0; step <= total; step++ {
		var line int
		if forward {
			line = (from.Line + step) % total
		} else {
			line = ((from.Line-step)%total + total) % total
		}
		matches := q.Matches(t.Line(line))
		if forward {
			for _, col := range matches {
				if step == 0 && col <= from.Col {
					continue
				}
				if step == total && col > from.Col {
					break
				}
				return buffer.Pos{Line: line, Col: col}, line < from.Line || (line == from.Line && step > 0), true
			}
		} else {
			for i := len(matches) - 1; i >= 0; i-- {
				col := matches[i]
				if step == 0 && col >= from.Col {
					continue
				}
				if step == total && col < from.Col {
					break
				}
				return buffer.Pos{Line: line, Col: col}, line > from.Line || (line == from.Line && step > 0), true
			}
		}
	}
	return from, false, false
}

func WordUnderCursor(line []rune, col int) ([]rune, int, bool) {
	if len(line) == 0 {
		return nil, 0, false
	}
	col = min(col, len(line)-1)
	for col < len(line) && charClass(line[col], false) != 2 {
		col++
	}
	if col >= len(line) {
		return nil, 0, false
	}
	start, end := col, col+1
	for start > 0 && charClass(line[start-1], false) == 2 {
		start--
	}
	for end < len(line) && charClass(line[end], false) == 2 {
		end++
	}
	return line[start:end], start, true
}
