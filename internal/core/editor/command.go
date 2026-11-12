package editor

import (
	"fmt"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

const msgUnsaved = "E37: alterações não salvas (use :w para salvar ou :q! para sair sem salvar)"

func (e *Editor) enterCommand() {
	e.mode = ModeCommand
	e.cmdline = e.cmdline[:0]
}

func (e *Editor) handleCommandKey(k input.Key) {
	switch {
	case k.Code == input.KeyEsc || isCtrl(k, 'c'):
		e.mode = ModeNormal
	case k.Code == input.KeyEnter:
		cmd := strings.TrimSpace(string(e.cmdline))
		e.mode = ModeNormal
		e.execute(cmd)
	case k.Code == input.KeyBackspace:
		if len(e.cmdline) == 0 {
			e.mode = ModeNormal
			return
		}
		e.cmdline = e.cmdline[:len(e.cmdline)-1]
	case k.Code == input.KeyRune && k.Mod == 0 && k.Rune != 0:
		e.cmdline = append(e.cmdline, k.Rune)
	}
}

func (e *Editor) execute(cmd string) {
	switch cmd {
	case "":
	case "w":
		e.write()
	case "q":
		e.requestQuit()
	case "q!":
		e.quit = true
	case "wq":
		if e.write() {
			e.quit = true
		}
	case "x":
		if !e.buf.Modified() || e.write() {
			e.quit = true
		}
	default:
		e.message = "E492: comando desconhecido: " + cmd
	}
}

func (e *Editor) requestQuit() {
	if e.buf.Modified() {
		e.message = msgUnsaved
		return
	}
	e.quit = true
}

func (e *Editor) write() bool {
	if e.path == "" {
		e.message = "E32: arquivo sem nome"
		return false
	}
	if e.buf.HasInvalidUTF8() {
		e.message = "arquivo tem bytes UTF-8 inválidos: salvar bloqueado para não corromper o original"
		return false
	}
	if err := e.save(e.path, e.buf); err != nil {
		e.message = "erro ao salvar: " + err.Error()
		return false
	}
	e.buf.MarkSaved()
	e.message = fmt.Sprintf("%q %dL gravado", e.filename, e.buf.LineCount())
	return true
}

func isCtrl(k input.Key, r rune) bool {
	return k.Code == input.KeyRune && k.Rune == r && k.Mod == input.ModCtrl
}
