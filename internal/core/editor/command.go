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
		if e.mode == ModeSearch {
			e.mode = ModeNormal
			e.executeSearch(string(e.cmdline))
			return
		}
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
	name, arg, _ := strings.Cut(cmd, " ")
	arg = strings.TrimSpace(arg)

	switch name {
	case "":
	case "noh", "nohlsearch":
		e.search.highlight = false
	case "e", "edit", "tabe", "tabedit", "tabnew":
		if arg == "" {
			e.message = "E32: informe o arquivo (:e caminho)"
			return
		}
		e.OpenFile(arg)
	case "w", "write":
		if arg != "" {
			e.message = "E492: :w com outro nome ainda não é suportado"
			return
		}
		e.write()
	case "wa", "wall":
		e.writeAll()
	case "q", "quit", "bd", "bdelete", "tabc", "tabclose":
		e.closeCurrent(false)
	case "q!", "quit!", "bd!", "bdelete!":
		e.closeCurrent(true)
	case "qa", "qall", "quitall":
		e.quitAll(false)
	case "qa!", "qall!", "quitall!":
		e.quitAll(true)
	case "wq", "x", "xit":
		if (name == "x" || name == "xit") && !e.buf.Modified() {
			e.closeCurrent(false)
			return
		}
		if e.write() {
			e.closeCurrent(false)
		}
	case "wqa", "wqall", "xa", "xall":
		if e.writeAll() {
			e.quit = true
		}
	case "bn", "bnext", "tabn", "tabnext":
		e.cycleTab(1)
	case "bp", "bprev", "bprevious", "tabp", "tabprev", "tabprevious":
		e.cycleTab(-1)
	default:
		e.message = "E492: comando desconhecido: " + cmd
	}
}

func (e *Editor) requestQuit() {
	e.quitAll(false)
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
	if e.lsp.client != nil && e.opened {
		e.lsp.client.Save(e.path)
	}
	e.message = fmt.Sprintf("%q %dL gravado", e.filename, e.buf.LineCount())
	return true
}

func isCtrl(k input.Key, r rune) bool {
	return k.Code == input.KeyRune && k.Rune == r && k.Mod == input.ModCtrl
}
