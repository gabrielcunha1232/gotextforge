package editor

import (
	"fmt"
	"math"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

type register struct {
	text     string
	linewise bool
	set      bool
}

func (e *Editor) handleNormalKey(k input.Key) {
	cmd, done := e.parser.Feed(k)
	if done {
		e.run(cmd)
	}
}

func (e *Editor) run(cmd engine.Command) {
	switch {
	case cmd.Action != engine.ActionNone:
		e.runAction(cmd)
	case cmd.Operator == engine.OpNone:
		e.runMotion(cmd)
	default:
		e.runOperator(cmd)
	}
}

func (e *Editor) motionContext(cmd engine.Command, operator bool) engine.MotionContext {
	return engine.MotionContext{
		Count:    cmd.Count,
		HasCount: cmd.Count > 0,
		Page:     e.viewport.Height,
		Operator: operator,
	}
}

func (e *Editor) runMotion(cmd engine.Command) {
	switch cmd.Motion.Kind {
	case engine.MotionTextObject, engine.MotionCurrentLine:
		return
	}
	t, ok := engine.ApplyMotion(e.buf, e.cursor, cmd.Motion, e.motionContext(cmd, false))
	if !ok {
		return
	}
	if t.Vertical {
		e.moveToLine(t.Pos.Line)
		return
	}
	e.cursor.Line = t.Pos.Line
	if cmd.Motion.Kind == engine.MotionLineEnd {
		e.cursor.Col = e.maxCol(t.Pos.Line)
		e.wantCol = math.MaxInt
		return
	}
	e.setCol(min(t.Pos.Col, e.maxCol(t.Pos.Line)))
}

func (e *Editor) operatorRange(cmd engine.Command) (engine.Range, bool) {
	m := cmd.Motion
	last := e.buf.LineCount() - 1

	switch m.Kind {
	case engine.MotionTextObject:
		return engine.TextObject(e.buf, e.cursor, m)
	case engine.MotionCurrentLine:
		end := min(e.cursor.Line+cmd.Times()-1, last)
		return engine.Range{
			Start:    buffer.Pos{Line: e.cursor.Line},
			End:      buffer.Pos{Line: end, Col: e.buf.LineLen(end)},
			Linewise: true,
		}, true
	}

	var t engine.Target
	ok := false
	isWord := m.Kind == engine.MotionWordStart || m.Kind == engine.MotionBigWordStart
	if cmd.Operator == engine.OpChange && isWord {
		t, ok = engine.ChangeWordTarget(e.buf, e.cursor, cmd.Times(), m.Kind == engine.MotionBigWordStart)
	}
	if !ok {
		t, ok = engine.ApplyMotion(e.buf, e.cursor, m, e.motionContext(cmd, true))
		if !ok {
			return engine.Range{}, false
		}
	}

	if t.Linewise {
		first, lastLine := min(e.cursor.Line, t.Pos.Line), max(e.cursor.Line, t.Pos.Line)
		return engine.Range{
			Start:    buffer.Pos{Line: first},
			End:      buffer.Pos{Line: lastLine, Col: e.buf.LineLen(lastLine)},
			Linewise: true,
		}, true
	}

	start, end := e.cursor, t.Pos
	if end.Before(start) {
		start, end = end, start
	}
	if t.Inclusive {
		end.Col++
	}
	if isWord && end.Line > start.Line && onlyBlankBefore(e.buf.Line(end.Line), end.Col) {
		prev := end.Line - 1
		end = buffer.Pos{Line: prev, Col: e.buf.LineLen(prev)}
	}
	return engine.Range{Start: e.buf.ClampPos(start), End: e.buf.ClampPos(end)}, true
}

func onlyBlankBefore(line []rune, col int) bool {
	for _, r := range line[:min(col, len(line))] {
		if r != ' ' && r != '\t' {
			return false
		}
	}
	return true
}

func (e *Editor) rangeText(r engine.Range) string {
	if r.Linewise {
		end := buffer.Pos{Line: r.End.Line, Col: e.buf.LineLen(r.End.Line)}
		return e.buf.Text(buffer.Pos{Line: r.Start.Line}, end) + "\n"
	}
	return e.buf.Text(r.Start, r.End)
}

func (e *Editor) runOperator(cmd engine.Command) {
	r, ok := e.operatorRange(cmd)
	if !ok {
		return
	}
	text := e.rangeText(r)
	e.reg = register{text: text, linewise: r.Linewise, set: true}

	switch cmd.Operator {
	case engine.OpYank:
		lines := r.End.Line - r.Start.Line + 1
		if r.Linewise && lines >= 3 {
			e.message = fmt.Sprintf("%d linhas copiadas", lines)
		}
		if r.Start.Before(e.cursor) {
			if r.Linewise {
				e.moveToLine(r.Start.Line)
			} else {
				e.cursor = r.Start
				e.setCol(min(r.Start.Col, e.maxCol(r.Start.Line)))
			}
		}

	case engine.OpDelete:
		if r.Linewise {
			e.deleteLines(r.Start.Line, r.End.Line)
			return
		}
		e.buf.Delete(r.Start, r.End)
		e.cursor = r.Start
		e.setCol(min(r.Start.Col, e.maxCol(r.Start.Line)))

	case engine.OpChange:
		if r.Linewise {
			end := buffer.Pos{Line: r.End.Line, Col: e.buf.LineLen(r.End.Line)}
			e.buf.Delete(buffer.Pos{Line: r.Start.Line}, end)
			e.enterInsert(buffer.Pos{Line: r.Start.Line})
			return
		}
		e.buf.Delete(r.Start, r.End)
		e.enterInsert(r.Start)
	}
}

func (e *Editor) deleteLines(first, last int) {
	lastLine := e.buf.LineCount() - 1
	target := first
	switch {
	case last < lastLine:
		e.buf.Delete(buffer.Pos{Line: first}, buffer.Pos{Line: last + 1})
	case first > 0:
		e.buf.Delete(buffer.Pos{Line: first - 1, Col: e.buf.LineLen(first - 1)}, buffer.Pos{Line: last, Col: e.buf.LineLen(last)})
		target = first - 1
	default:
		e.buf.Delete(buffer.Pos{}, buffer.Pos{Line: last, Col: e.buf.LineLen(last)})
		target = 0
	}
	e.cursor.Line = target
	e.setCol(engine.FirstNonBlank(e.buf.Line(target)))
	if n := last - first + 1; n >= 3 {
		e.message = fmt.Sprintf("%d linhas a menos", n)
	}
}

func (e *Editor) runAction(cmd engine.Command) {
	line := e.cursor.Line
	switch cmd.Action {
	case engine.ActionInsert:
		e.enterInsert(e.cursor)
	case engine.ActionAppend:
		e.enterInsert(buffer.Pos{Line: line, Col: min(e.cursor.Col+1, e.buf.LineLen(line))})
	case engine.ActionInsertLineStart:
		e.enterInsert(buffer.Pos{Line: line, Col: engine.FirstNonBlank(e.buf.Line(line))})
	case engine.ActionAppendLineEnd:
		e.enterInsert(buffer.Pos{Line: line, Col: e.buf.LineLen(line)})
	case engine.ActionOpenBelow:
		end := e.buf.Insert(buffer.Pos{Line: line, Col: e.buf.LineLen(line)}, "\n")
		e.enterInsert(end)
	case engine.ActionOpenAbove:
		e.buf.Insert(buffer.Pos{Line: line}, "\n")
		e.enterInsert(buffer.Pos{Line: line})
	case engine.ActionPasteAfter, engine.ActionPasteBefore:
		e.paste(cmd.Action == engine.ActionPasteAfter, cmd.Times())
	case engine.ActionUndo:
		for i := 0; i < cmd.Times(); i++ {
			pos, ok := e.buf.Undo()
			e.afterHistory(pos, ok, "Já na alteração mais antiga")
			if !ok {
				break
			}
		}
	case engine.ActionRedo:
		for i := 0; i < cmd.Times(); i++ {
			pos, ok := e.buf.Redo()
			e.afterHistory(pos, ok, "Já na alteração mais recente")
			if !ok {
				break
			}
		}
	case engine.ActionJoin:
		e.join(max(cmd.Times()-1, 1))
	case engine.ActionReplace:
		e.replace(cmd.Char, cmd.Times())
	case engine.ActionCommandLine:
		e.enterCommand()
	case engine.ActionQuit:
		e.requestQuit()
	case engine.ActionHover:
		e.requestHover()
	}
}

func (e *Editor) paste(after bool, count int) {
	if !e.reg.set {
		e.message = "E353: nada no registrador"
		return
	}
	text := strings.Repeat(e.reg.text, count)
	line := e.cursor.Line

	if e.reg.linewise {
		target := line
		if after {
			e.buf.Insert(buffer.Pos{Line: line, Col: e.buf.LineLen(line)}, "\n"+strings.TrimSuffix(text, "\n"))
			target = line + 1
		} else {
			e.buf.Insert(buffer.Pos{Line: line}, text)
		}
		e.cursor.Line = target
		e.setCol(engine.FirstNonBlank(e.buf.Line(target)))
		return
	}

	at := e.cursor
	if after && e.buf.LineLen(line) > 0 {
		at.Col++
	}
	end := e.buf.Insert(at, text)
	e.cursor = end
	e.setCol(max(end.Col-1, 0))
}

func (e *Editor) join(times int) {
	for ; times > 0; times-- {
		line := e.cursor.Line
		if line >= e.buf.LineCount()-1 {
			return
		}
		cur := e.buf.Line(line)
		next := e.buf.Line(line + 1)
		lead := 0
		for lead < len(next) && (next[lead] == ' ' || next[lead] == '\t') {
			lead++
		}
		at := buffer.Pos{Line: line, Col: len(cur)}
		e.buf.Delete(at, buffer.Pos{Line: line + 1, Col: lead})
		rest := next[lead:]
		if len(cur) > 0 && len(rest) > 0 && rest[0] != ')' {
			e.buf.Insert(at, " ")
		}
		e.cursor = at
		e.setCol(min(at.Col, e.maxCol(line)))
	}
}

func (e *Editor) replace(r rune, count int) {
	line := e.cursor.Line
	if e.cursor.Col+count > e.buf.LineLen(line) {
		return
	}
	end := buffer.Pos{Line: line, Col: e.cursor.Col + count}
	e.buf.Delete(e.cursor, end)
	e.buf.Insert(e.cursor, strings.Repeat(string(r), count))
	e.setCol(e.cursor.Col + count - 1)
}

func (e *Editor) handleInsertMotion(k input.Key) bool {
	switch k.Code {
	case input.KeyDown:
		e.moveVertical(1)
	case input.KeyUp:
		e.moveVertical(-1)
	case input.KeyLeft:
		e.moveHorizontal(-1)
	case input.KeyRight:
		e.moveHorizontal(1)
	case input.KeyPageDown:
		e.moveVertical(e.viewport.Height)
	case input.KeyPageUp:
		e.moveVertical(-e.viewport.Height)
	case input.KeyHome:
		e.setCol(0)
	case input.KeyEnd:
		e.cursor.Col = e.maxCol(e.cursor.Line)
		e.wantCol = math.MaxInt
	default:
		return false
	}
	return true
}

func (e *Editor) afterHistory(pos buffer.Pos, ok bool, emptyMsg string) {
	if !ok {
		e.message = emptyMsg
		return
	}
	e.cursor.Line = pos.Line
	e.setCol(min(pos.Col, e.maxCol(pos.Line)))
}
