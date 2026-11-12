package editor

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/fuzzy"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

const (
	finderMaxFiles   = 50000
	finderMaxResults = 300
)

var skipDirs = map[string]bool{".git": true, "node_modules": true, ".idea": true, ".vscode": true}

type finder struct {
	title       string
	items       []string
	targets     map[string]string
	query       []rune
	matches     []fuzzy.Match
	sel         int
	allowCreate bool
	truncated   bool
}

func (f *finder) createRow() bool {
	if !f.allowCreate || len(f.query) == 0 {
		return false
	}
	q := strings.TrimSpace(string(f.query))
	for _, m := range f.matches {
		if strings.EqualFold(m.Text, q) {
			return false
		}
	}
	return q != ""
}

func (f *finder) rowCount() int {
	n := len(f.matches)
	if f.createRow() {
		n++
	}
	return n
}

func (f *finder) refilter() {
	f.matches = fuzzy.Filter(f.items, string(f.query), finderMaxResults)
	f.sel = min(f.sel, max(f.rowCount()-1, 0))
}

func (e *Editor) openFileFinder() {
	root := e.projectRoot()
	f := &finder{title: " Arquivos ", targets: map[string]string{}, allowCreate: true}
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if len(f.items) >= finderMaxFiles {
			f.truncated = true
			return filepath.SkipAll
		}
		rel := e.relToRoot(path)
		f.items = append(f.items, rel)
		f.targets[rel] = path
		return nil
	})
	f.refilter()
	e.finder = f
	e.mode = ModeFinder
}

func (e *Editor) openBufferFinder() {
	f := &finder{title: " Abas abertas ", targets: map[string]string{}}
	for _, d := range e.docs {
		if d.path == "" {
			continue
		}
		label := e.relToRoot(d.path)
		if d.buf.Modified() {
			label += " [+]"
		}
		f.items = append(f.items, label)
		f.targets[label] = d.path
	}
	if len(f.items) == 0 {
		e.message = "nenhum arquivo aberto"
		return
	}
	f.refilter()
	e.finder = f
	e.mode = ModeFinder
}

func (e *Editor) promptNewFile() {
	initial := ""
	if e.path != "" {
		if dir := e.relToRoot(filepath.Dir(e.path)); dir != "" && !strings.HasPrefix(dir, "..") && !filepath.IsAbs(dir) {
			initial = dir + "/"
		}
	}
	e.openPrompt("novo arquivo: ", initial, func(name string) {
		e.createPath(e.projectRoot(), name, true)
	})
}

func (e *Editor) handleFinderKey(k input.Key) {
	f := e.finder
	if f == nil {
		e.mode = ModeNormal
		return
	}
	n := f.rowCount()
	switch {
	case k.Code == input.KeyEsc || isCtrl(k, 'c'):
		e.finder = nil
		e.mode = ModeNormal
	case k.Code == input.KeyEnter:
		e.finder = nil
		e.mode = ModeNormal
		switch {
		case f.sel < len(f.matches):
			e.OpenFile(f.targets[f.matches[f.sel].Text])
		case f.createRow():
			e.createPath(e.projectRoot(), strings.TrimSpace(string(f.query)), true)
		}
	case k.Code == input.KeyDown || isCtrl(k, 'n') || isCtrl(k, 'j') || k.Code == input.KeyTab:
		if n > 0 {
			f.sel = (f.sel + 1) % n
		}
	case k.Code == input.KeyUp || isCtrl(k, 'p') || isCtrl(k, 'k'):
		if n > 0 {
			f.sel = (f.sel - 1 + n) % n
		}
	case k.Code == input.KeyBackspace:
		if len(f.query) > 0 {
			f.query = f.query[:len(f.query)-1]
			f.sel = 0
			f.refilter()
		}
	case isCtrl(k, 'u'):
		f.query = f.query[:0]
		f.sel = 0
		f.refilter()
	case k.Code == input.KeyRune && k.Mod == 0 && k.Rune != 0:
		f.query = append(f.query, k.Rune)
		f.sel = 0
		f.refilter()
	}
}

func (e *Editor) finderBox(rows int) ([]render.StyledLine, int, int) {
	f := e.finder
	w := min(e.width-4, 90)
	h := min(rows-2, 22)
	if w < 20 || h < 5 {
		w, h = max(e.width, 1), max(rows, 1)
	}
	inner := w - 4
	listRows := h - 4

	border := render.Style{Fg: colBorder, Bg: colBgFloat}
	fill := render.Style{Bg: colBgFloat}
	title := f.title
	top := "╭─" + title + strings.Repeat("─", max(w-3-len([]rune(title)), 0)) + "╮"
	box := []render.StyledLine{{{Text: top, Style: render.Style{Fg: colBorder, Bg: colBgFloat, Bold: true}}}}

	count := fmt.Sprintf("%d/%d", len(f.matches), len(f.items))
	if f.truncated {
		count += "+"
	}
	q := sanitize(string(f.query))
	promptLine := render.StyledLine{
		{Text: "> ", Style: render.Style{Fg: colOrange, Bg: colBgFloat, Bold: true}},
		{Text: q, Style: render.Style{Fg: colFg, Bg: colBgFloat}},
	}
	gap := inner - promptLine.Width() - len(count)
	if gap > 0 {
		promptLine = append(promptLine,
			render.Span{Text: strings.Repeat(" ", gap), Style: fill},
			render.Span{Text: count, Style: render.Style{Fg: colComment, Bg: colBgFloat}})
	}
	box = append(box, frameRow(promptLine, inner, border, fill))
	box = append(box, render.StyledLine{{Text: "├" + strings.Repeat("─", w-2) + "┤", Style: border}})

	first := 0
	if f.sel >= listRows {
		first = f.sel - listRows + 1
	}
	total := f.rowCount()
	for i := first; i < first+listRows; i++ {
		var line render.StyledLine
		bg := colBgFloat
		if i == f.sel {
			bg = colBgSel
		}
		switch {
		case i < len(f.matches):
			line = finderRow(f.matches[i], i == f.sel, bg)
		case i < total:
			line = render.StyledLine{
				{Text: "  ", Style: render.Style{Bg: bg}},
				{Text: "+ criar " + sanitize(strings.TrimSpace(string(f.query))), Style: render.Style{Fg: colGreen, Bg: bg, Bold: i == f.sel}},
			}
		}
		box = append(box, frameRow(line, inner, border, render.Style{Bg: bg}))
	}
	box = append(box, render.StyledLine{{Text: "╰" + strings.Repeat("─", w-2) + "╯", Style: border}})

	row := max((rows-len(box))/2, 0)
	col := max((e.width-w)/2, 0)
	return box, row, col
}

func frameRow(content render.StyledLine, inner int, border, fill render.Style) render.StyledLine {
	row := render.StyledLine{{Text: "│ ", Style: border}}
	row = append(row, padTo(truncateLine(content, inner), inner, fill)...)
	return append(row, render.Span{Text: " │", Style: border})
}

func finderRow(m fuzzy.Match, selected bool, bg render.Color) render.StyledLine {
	runes := []rune(sanitize(m.Text))
	nameStart := 0
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '/' {
			nameStart = i + 1
			break
		}
	}
	hit := map[int]bool{}
	for _, p := range m.Positions {
		hit[p] = true
	}
	marker := "  "
	if selected {
		marker = "▶ "
	}
	line := render.StyledLine{{Text: marker, Style: render.Style{Fg: colOrange, Bg: bg}}}
	for i, r := range runes {
		style := render.Style{Fg: colComment, Bg: bg}
		if i >= nameStart {
			style = render.Style{Fg: colFg, Bg: bg, Bold: selected}
		}
		if hit[i] {
			style = render.Style{Fg: colOrange, Bg: bg, Bold: true}
		}
		line = append(line, render.Span{Text: string(r), Style: style})
	}
	return line
}
