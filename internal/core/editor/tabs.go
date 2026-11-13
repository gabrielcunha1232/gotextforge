package editor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func loadFile(path string) (*buffer.Buffer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return buffer.Load(f)
}

func (e *Editor) tablineRows() int {
	if len(e.docs) > 1 {
		return 1
	}
	return 0
}

func (e *Editor) textHeight() int {
	return max(e.height-1-e.tablineRows()-e.terminalRows(), 1)
}

func (e *Editor) layout() {
	h := e.textHeight()
	for _, d := range e.docs {
		d.viewport.Resize(h, d.buf.LineCount())
		d.viewport.EnsureVisible(d.cursor.Line, d.buf.LineCount())
	}
}

func samePath(a, b string) bool {
	ca, errA := filepath.Abs(a)
	cb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		ca, cb = filepath.Clean(a), filepath.Clean(b)
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ca, cb)
	}
	return ca == cb
}

func (e *Editor) isPristineIntro() bool {
	return len(e.docs) == 1 && e.path == "" && !e.buf.Modified() &&
		e.buf.LineCount() == 1 && e.buf.LineLen(0) == 0
}

func (e *Editor) OpenFile(path string) {
	switch e.mode {
	case ModeExplorer, ModeFinder, ModePrompt:
		e.mode = ModeNormal
	}
	for i, d := range e.docs {
		if d.path != "" && samePath(d.path, path) {
			e.switchTo(i)
			return
		}
	}

	buf, err := e.load(path)
	isNew := false
	switch {
	case errors.Is(err, fs.ErrNotExist):
		buf, isNew = buffer.New(), true
	case err != nil:
		e.message = fmt.Sprintf("erro ao abrir %q: %v", path, err)
		return
	}

	doc := newDocument(buf, path)
	if e.isPristineIntro() {
		e.docs[0] = doc
	} else {
		e.docs = append(e.docs, doc)
	}
	e.layout()
	e.switchTo(len(e.docs) - 1)
	if isNew {
		e.message = fmt.Sprintf("%q [Novo]", doc.filename)
	} else {
		e.message = fmt.Sprintf("%q %dL", doc.filename, buf.LineCount())
	}
}

func (e *Editor) switchTo(i int) {
	if i < 0 || i >= len(e.docs) {
		return
	}
	e.document = e.docs[i]
	e.hover = nil
	e.comp = nil
	e.ensureLanguageServer(e.document)
}

func (e *Editor) currentIndex() int {
	for i, d := range e.docs {
		if d == e.document {
			return i
		}
	}
	return 0
}

func (e *Editor) cycleTab(delta int) {
	n := len(e.docs)
	if n < 2 {
		return
	}
	e.switchTo(((e.currentIndex()+delta)%n + n) % n)
}

func (e *Editor) closeCurrent(force bool) {
	if e.buf.Modified() && !force {
		e.message = msgUnsaved
		return
	}
	if len(e.docs) == 1 {
		e.quit = true
		return
	}
	i := e.currentIndex()
	if e.lsp.client != nil && e.opened {
		e.lsp.client.CloseDocument(e.path)
	}
	e.docs = append(e.docs[:i], e.docs[i+1:]...)
	e.layout()
	e.switchTo(min(i, len(e.docs)-1))
}

func (e *Editor) firstModified() *document {
	for _, d := range e.docs {
		if d.buf.Modified() {
			return d
		}
	}
	return nil
}

func (e *Editor) quitAll(force bool) {
	if d := e.firstModified(); d != nil && !force {
		e.message = fmt.Sprintf("E37: %q tem alterações não salvas (use :wa para salvar tudo ou :qa! para sair sem salvar)", d.filename)
		return
	}
	e.quit = true
}

func (e *Editor) writeAll() bool {
	current := e.document
	defer func() { e.document = current }()
	written := 0
	for _, d := range e.docs {
		if !d.buf.Modified() {
			continue
		}
		e.document = d
		if !e.write() {
			return false
		}
		written++
	}
	e.message = fmt.Sprintf("%d arquivo(s) gravado(s)", written)
	return true
}

func (e *Editor) tabline() render.StyledLine {
	if len(e.docs) < 2 {
		return nil
	}
	active := render.Style{Fg: colFg, Bg: colBgHighlight, Bold: true}
	inactive := render.Style{Fg: colDark3, Bg: colBgFloat}
	marker := render.Style{Fg: colBlue, Bg: colBgHighlight}

	segments := make([]render.StyledLine, len(e.docs))
	cur := e.currentIndex()
	for i, d := range e.docs {
		name := sanitize(d.filename)
		mod := ""
		if d.buf.Modified() {
			mod = " ●"
		}
		if i == cur {
			segments[i] = render.StyledLine{
				{Text: "▎", Style: marker},
				{Text: fmt.Sprintf("%d %s%s ", i+1, name, mod), Style: active},
			}
		} else {
			segments[i] = render.StyledLine{{Text: fmt.Sprintf(" %d %s%s ", i+1, name, mod), Style: inactive}}
		}
	}

	first := 0
	for first < cur {
		w := 0
		for _, s := range segments[first : cur+1] {
			w += s.Width()
		}
		if w <= e.width {
			break
		}
		first++
	}

	var line render.StyledLine
	for _, s := range segments[first:] {
		line = append(line, s...)
	}
	return padTo(truncateLine(line, e.width), e.width, render.Style{Bg: colBgFloat})
}

func (e *Editor) SwitchToFirst() {
	e.switchTo(0)
}
