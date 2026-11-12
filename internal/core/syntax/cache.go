package syntax

type LineSource interface {
	LineCount() int
	Line(i int) []rune
}

type Cache struct {
	states []State
}

func (c *Cache) StateAt(src LineSource, line int) State {
	if len(c.states) == 0 {
		c.states = append(c.states, StateNormal)
	}
	last := src.LineCount() - 1
	if line > last {
		line = last
	}
	for len(c.states) <= line {
		i := len(c.states) - 1
		c.states = append(c.states, NextState(src.Line(i), c.states[i]))
	}
	return c.states[line]
}

func NextState(line []rune, in State) State {
	switch in {
	case StateNormal:
		if !containsPair(line, '/', '*') && !containsRune(line, '`') {
			return StateNormal
		}
	case StateBlockComment:
		if !containsPair(line, '*', '/') {
			return StateBlockComment
		}
	case StateRawString:
		if !containsRune(line, '`') {
			return StateRawString
		}
	}
	_, out := HighlightLine(line, in)
	return out
}

func containsRune(line []rune, r rune) bool {
	for _, c := range line {
		if c == r {
			return true
		}
	}
	return false
}

func containsPair(line []rune, a, b rune) bool {
	for i := 0; i+1 < len(line); i++ {
		if line[i] == a && line[i+1] == b {
			return true
		}
	}
	return false
}

func (c *Cache) Invalidate(fromLine int) {
	keep := fromLine + 1
	if keep < 0 {
		keep = 0
	}
	if keep < len(c.states) {
		c.states = c.states[:keep]
	}
}
