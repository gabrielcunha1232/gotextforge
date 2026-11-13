package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func (e *Editor) enterInsert(at buffer.Pos) {
	e.mode = ModeInsert
	e.cursor = e.buf.ClampPos(at)
	e.wantCol = e.cursor.Col
}

func (e *Editor) leaveInsert() {
	e.comp = nil
	e.buf.EndChange()
	e.mode = ModeNormal
	if e.cursor.Col > 0 {
		e.cursor.Col--
	}
	e.setCol(min(e.cursor.Col, e.maxCol(e.cursor.Line)))
}

func (e *Editor) handleInsertKey(k input.Key) {
	if e.handleCompletionKey(k) {
		return
	}
	if e.handleInsertMotion(k) {
		return
	}

	switch {
	case k.Code == input.KeyEsc || isCtrl(k, 'c'):
		e.leaveInsert()
	case k.Code == input.KeyEnter && k.Mod == 0:
		e.newline()
	case k.Code == input.KeyTab && k.Mod == 0:
		e.insertText("\t")
	case k.Code == input.KeyBackspace:
		if !e.backspacePair() {
			e.backspace()
		}
	case k.Code == input.KeyDelete:
		e.deleteForward()
	case k.Code == input.KeyRune && k.Mod == 0 && k.Rune != 0:
		e.typeRune(k.Rune)
	}
}

func (e *Editor) insertText(text string) {
	e.cursor = e.buf.Insert(e.cursor, text)
	e.wantCol = e.cursor.Col
}

func (e *Editor) backspace() {
	switch {
	case e.cursor.Col > 0:
		start := buffer.Pos{Line: e.cursor.Line, Col: e.cursor.Col - 1}
		e.buf.Delete(start, e.cursor)
		e.cursor = start
	case e.cursor.Line > 0:
		prev := e.cursor.Line - 1
		start := buffer.Pos{Line: prev, Col: e.buf.LineLen(prev)}
		e.buf.Delete(start, e.cursor)
		e.cursor = start
	}
	e.wantCol = e.cursor.Col
}

func (e *Editor) deleteForward() {
	end := buffer.Pos{Line: e.cursor.Line, Col: e.cursor.Col + 1}
	if e.cursor.Col >= e.buf.LineLen(e.cursor.Line) {
		if e.cursor.Line >= e.buf.LineCount()-1 {
			return
		}
		end = buffer.Pos{Line: e.cursor.Line + 1, Col: 0}
	}
	e.buf.Delete(e.cursor, end)
}
