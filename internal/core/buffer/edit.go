package buffer

import (
	"slices"
	"strings"
)

func (b *Buffer) ClampPos(p Pos) Pos {
	if p.Line < 0 {
		p.Line = 0
	}
	if last := len(b.lines) - 1; p.Line > last {
		p.Line = last
	}
	if p.Col < 0 {
		p.Col = 0
	}
	if n := len(b.lines[p.Line]); p.Col > n {
		p.Col = n
	}
	return p
}

func (b *Buffer) Insert(p Pos, text string) Pos {
	p = b.ClampPos(p)
	if text == "" {
		return p
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n")
	line := b.lines[p.Line]

	if len(parts) == 1 {
		ins := []rune(parts[0])
		newLine := make([]rune, 0, len(line)+len(ins))
		newLine = append(newLine, line[:p.Col]...)
		newLine = append(newLine, ins...)
		newLine = append(newLine, line[p.Col:]...)
		b.lines[p.Line] = newLine
		b.modified = true
		return Pos{Line: p.Line, Col: p.Col + len(ins)}
	}

	newLines := make([][]rune, len(parts))

	first := []rune(parts[0])
	newLines[0] = append(append(make([]rune, 0, p.Col+len(first)), line[:p.Col]...), first...)

	for i := 1; i < len(parts)-1; i++ {
		newLines[i] = []rune(parts[i])
	}

	lastIns := []rune(parts[len(parts)-1])
	tail := line[p.Col:]
	newLines[len(parts)-1] = append(append(make([]rune, 0, len(lastIns)+len(tail)), lastIns...), tail...)

	b.lines = slices.Replace(b.lines, p.Line, p.Line+1, newLines...)
	b.modified = true
	return Pos{Line: p.Line + len(parts) - 1, Col: len(lastIns)}
}

func (b *Buffer) Delete(start, end Pos) string {
	start, end = b.ClampPos(start), b.ClampPos(end)
	if end.Before(start) {
		start, end = end, start
	}
	if start == end {
		return ""
	}

	removed := b.Text(start, end)

	head := b.lines[start.Line][:start.Col]
	tail := b.lines[end.Line][end.Col:]
	joined := make([]rune, 0, len(head)+len(tail))
	joined = append(joined, head...)
	joined = append(joined, tail...)

	b.lines = slices.Replace(b.lines, start.Line, end.Line+1, joined)
	b.modified = true
	return removed
}

func (b *Buffer) Text(start, end Pos) string {
	start, end = b.ClampPos(start), b.ClampPos(end)
	if end.Before(start) {
		start, end = end, start
	}

	if start.Line == end.Line {
		return string(b.lines[start.Line][start.Col:end.Col])
	}

	var sb strings.Builder
	sb.WriteString(string(b.lines[start.Line][start.Col:]))
	for i := start.Line + 1; i < end.Line; i++ {
		sb.WriteByte('\n')
		sb.WriteString(string(b.lines[i]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(b.lines[end.Line][:end.Col]))
	return sb.String()
}

func (b *Buffer) Modified() bool {
	return b.modified
}

func (b *Buffer) MarkSaved() {
	b.modified = false
}
