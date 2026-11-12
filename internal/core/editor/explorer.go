package editor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type treeNode struct {
	name     string
	path     string
	dir      bool
	expanded bool
	loaded   bool
	depth    int
	parent   *treeNode
	children []*treeNode
}

type explorer struct {
	open       bool
	tree       *treeNode
	rows       []*treeNode
	sel        int
	top        int
	showHidden bool
}

func (e *Editor) SetRoot(dir string) {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	e.root = dir
	if e.explorer.tree != nil {
		e.buildTree()
	}
}

func (e *Editor) projectRoot() string {
	if e.root == "" {
		if wd, err := os.Getwd(); err == nil {
			e.root = wd
		} else {
			e.root = "."
		}
	}
	return e.root
}

func (e *Editor) buildTree() {
	root := e.projectRoot()
	e.explorer.tree = &treeNode{name: filepath.Base(root), path: root, dir: true, expanded: true, depth: -1}
	e.loadChildren(e.explorer.tree)
	e.explorer.sel, e.explorer.top = 0, 0
	e.flatten()
}

func (e *Editor) loadChildren(n *treeNode) {
	entries, err := os.ReadDir(n.path)
	n.loaded = true
	n.children = nil
	if err != nil {
		e.message = fmt.Sprintf("erro ao ler %q: %v", n.path, err)
		return
	}
	for _, ent := range entries {
		if !e.explorer.showHidden && strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		n.children = append(n.children, &treeNode{
			name:   ent.Name(),
			path:   filepath.Join(n.path, ent.Name()),
			dir:    ent.IsDir(),
			depth:  n.depth + 1,
			parent: n,
		})
	}
	sort.SliceStable(n.children, func(i, j int) bool {
		a, b := n.children[i], n.children[j]
		if a.dir != b.dir {
			return a.dir
		}
		return strings.ToLower(a.name) < strings.ToLower(b.name)
	})
}

func (e *Editor) flatten() {
	var selPath string
	if s := e.selectedNode(); s != nil {
		selPath = s.path
	}
	e.explorer.rows = e.explorer.rows[:0]
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		for _, c := range n.children {
			e.explorer.rows = append(e.explorer.rows, c)
			if c.dir && c.expanded {
				walk(c)
			}
		}
	}
	if e.explorer.tree != nil {
		walk(e.explorer.tree)
	}
	e.selectPath(selPath)
}

func (e *Editor) selectedNode() *treeNode {
	x := &e.explorer
	if x.sel >= 0 && x.sel < len(x.rows) {
		return x.rows[x.sel]
	}
	return nil
}

func (e *Editor) selectPath(path string) bool {
	for i, r := range e.explorer.rows {
		if path != "" && samePath(r.path, path) {
			e.explorer.sel = i
			return true
		}
	}
	e.explorer.sel = min(e.explorer.sel, max(len(e.explorer.rows)-1, 0))
	return false
}

func (e *Editor) reveal(path string) {
	if path == "" || e.explorer.tree == nil {
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	rel, err := filepath.Rel(e.explorer.tree.path, abs)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return
	}
	node := e.explorer.tree
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		var next *treeNode
		for _, c := range node.children {
			if c.dir && strings.EqualFold(c.name, part) {
				next = c
				break
			}
		}
		if next == nil {
			return
		}
		if !next.loaded {
			e.loadChildren(next)
		}
		next.expanded = true
		node = next
	}
	e.flatten()
	e.selectPath(abs)
}

func (e *Editor) openExplorer(focus bool) {
	if e.explorer.tree == nil || !samePath(e.explorer.tree.path, e.projectRoot()) {
		e.buildTree()
	}
	e.explorer.open = true
	e.reveal(e.path)
	if focus {
		e.mode = ModeExplorer
	}
}

func (e *Editor) toggleExplorer() {
	switch {
	case !e.explorer.open:
		e.openExplorer(true)
	case e.mode == ModeExplorer:
		e.explorer.open = false
		e.mode = ModeNormal
	default:
		e.openExplorer(true)
	}
}

func (e *Editor) explorerTargetDir() string {
	n := e.selectedNode()
	switch {
	case n == nil:
		return e.projectRoot()
	case n.dir:
		return n.path
	default:
		return filepath.Dir(n.path)
	}
}

func (e *Editor) relToRoot(path string) string {
	rel, err := filepath.Rel(e.projectRoot(), path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	if rel == "." {
		return ""
	}
	return filepath.ToSlash(rel)
}

func (e *Editor) handleExplorerKey(k input.Key) {
	x := &e.explorer
	if k.Code == input.KeyRune && k.Rune == ' ' || e.parser.Pending() != "" {
		if cmd, done := e.parser.Feed(k); done && cmd.Action != engine.ActionNone {
			e.runAction(cmd)
		}
		return
	}

	n := e.selectedNode()
	last := len(x.rows) - 1
	plain := func(r rune) bool { return k.Code == input.KeyRune && k.Rune == r && k.Mod == 0 }

	switch {
	case plain('j') || k.Code == input.KeyDown:
		x.sel = min(x.sel+1, max(last, 0))
	case plain('k') || k.Code == input.KeyUp:
		x.sel = max(x.sel-1, 0)
	case plain('g') || k.Code == input.KeyHome:
		x.sel = 0
	case plain('G') || k.Code == input.KeyEnd:
		x.sel = max(last, 0)
	case k.Code == input.KeyPageDown:
		x.sel = min(x.sel+e.textHeight()/2, max(last, 0))
	case k.Code == input.KeyPageUp:
		x.sel = max(x.sel-e.textHeight()/2, 0)
	case plain('l') || k.Code == input.KeyRight:
		e.explorerOpenNode(n, false)
	case plain('o') || k.Code == input.KeyEnter:
		e.explorerOpenNode(n, true)
	case plain('h') || k.Code == input.KeyLeft:
		switch {
		case n == nil:
		case n.dir && n.expanded:
			n.expanded = false
			e.flatten()
		case n.parent != nil && n.parent != x.tree:
			e.selectPath(n.parent.path)
		}
	case k.Code == input.KeyBackspace || plain('-'):
		old := x.tree.path
		parent := filepath.Dir(old)
		if parent == old {
			e.message = "já está na raiz do disco"
			return
		}
		e.SetRoot(parent)
		e.selectPath(old)
		e.message = "pasta: " + parent
	case plain('.'):
		dir := e.explorerTargetDir()
		e.SetRoot(dir)
		e.message = "pasta do projeto: " + dir
	case plain('H'):
		x.showHidden = !x.showHidden
		e.refreshTree()
	case plain('R'):
		e.refreshTree()
	case plain('a'):
		dir := e.explorerTargetDir()
		label := fmt.Sprintf("novo em %s/ (termine com / para pasta): ", e.relToRoot(dir))
		e.openPrompt(label, "", func(name string) { e.createPath(dir, name, true) })
	case plain('r') && n != nil:
		e.openPrompt("renomear para: ", n.name, func(name string) { e.renamePath(n, name) })
	case plain('d') && n != nil:
		e.openPrompt(fmt.Sprintf("apagar %q? (s/N): ", n.name), "", func(ans string) {
			if a := strings.ToLower(strings.TrimSpace(ans)); a == "s" || a == "sim" || a == "y" {
				e.deletePath(n)
			}
		})
	case plain('q'):
		x.open = false
		e.mode = ModeNormal
	case k.Code == input.KeyEsc:
		e.mode = ModeNormal
	}
}

func (e *Editor) explorerOpenNode(n *treeNode, toggle bool) {
	if n == nil {
		return
	}
	if n.dir {
		switch {
		case !n.expanded:
			if !n.loaded {
				e.loadChildren(n)
			}
			n.expanded = true
		case toggle:
			n.expanded = false
		case len(n.children) > 0:
			e.flatten()
			e.selectPath(n.children[0].path)
			return
		}
		e.flatten()
		return
	}
	e.mode = ModeNormal
	e.OpenFile(n.path)
}

func (e *Editor) refreshTree() {
	if e.explorer.tree == nil {
		return
	}
	var reload func(n *treeNode)
	reload = func(n *treeNode) {
		expanded := map[string]bool{}
		for _, c := range n.children {
			if c.dir && c.expanded {
				expanded[c.name] = true
			}
		}
		e.loadChildren(n)
		for _, c := range n.children {
			if expanded[c.name] {
				c.expanded = true
				reload(c)
			}
		}
	}
	reload(e.explorer.tree)
	e.flatten()
}

func (e *Editor) createPath(dir, name string, open bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	isDir := strings.HasSuffix(name, "/") || strings.HasSuffix(name, "\\")
	path := filepath.Join(dir, filepath.FromSlash(name))
	if _, err := os.Stat(path); err == nil {
		e.message = fmt.Sprintf("%q já existe", e.relToRoot(path))
		if !isDir && open {
			e.OpenFile(path)
		}
		return
	}

	var err error
	if isDir {
		err = os.MkdirAll(path, 0o755)
	} else if err = os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		var f *os.File
		if f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644); err == nil {
			err = f.Close()
		}
	}
	if err != nil {
		e.message = "erro ao criar: " + err.Error()
		return
	}

	e.refreshTree()
	e.reveal(path)
	kind := "arquivo"
	if isDir {
		kind = "pasta"
	}
	if !isDir && open {
		e.mode = ModeNormal
		e.OpenFile(path)
	}
	e.message = fmt.Sprintf("%s criado: %s", kind, e.relToRoot(path))
}

func (e *Editor) renamePath(n *treeNode, name string) {
	name = strings.TrimSpace(name)
	if name == "" || name == n.name {
		return
	}
	newPath := filepath.Join(filepath.Dir(n.path), filepath.FromSlash(name))
	if _, err := os.Stat(newPath); err == nil {
		e.message = fmt.Sprintf("%q já existe", name)
		return
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		e.message = "erro ao renomear: " + err.Error()
		return
	}
	if err := os.Rename(n.path, newPath); err != nil {
		e.message = "erro ao renomear: " + err.Error()
		return
	}
	for _, d := range e.docs {
		if d.path == "" {
			continue
		}
		rel, err := filepath.Rel(n.path, d.path)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		if e.lsp.client != nil && d.opened {
			e.lsp.client.CloseDocument(d.path)
			d.opened = false
		}
		d.path = filepath.Join(newPath, rel)
		d.filename = filepath.Base(d.path)
		d.highlight = strings.HasSuffix(d.path, ".go")
	}
	e.ensureLanguageServer(e.document)
	e.refreshTree()
	e.reveal(newPath)
	e.message = "renomeado para " + e.relToRoot(newPath)
}

func (e *Editor) deletePath(n *treeNode) {
	for _, d := range e.docs {
		if d.path != "" && samePath(d.path, n.path) && d.buf.Modified() {
			e.message = fmt.Sprintf("%q tem alterações não salvas: salve ou descarte antes de apagar", n.name)
			return
		}
	}
	if err := os.Remove(n.path); err != nil {
		if n.dir && !errors.Is(err, fs.ErrNotExist) {
			e.message = fmt.Sprintf("não apaguei %q: só pastas vazias podem ser apagadas aqui", n.name)
		} else {
			e.message = "erro ao apagar: " + err.Error()
		}
		return
	}
	for i := 0; i < len(e.docs); i++ {
		if d := e.docs[i]; d.path != "" && samePath(d.path, n.path) {
			current := e.document
			e.document = d
			e.closeCurrent(true)
			if current != d {
				e.switchTo(e.indexOf(current))
			}
			i--
		}
	}
	if e.quit {
		e.quit = false
		e.docs = []*document{newDocument(nil, "")}
		e.document = e.docs[0]
		e.layout()
	}
	e.refreshTree()
	e.message = "apagado: " + n.name
}

func (e *Editor) indexOf(d *document) int {
	for i, x := range e.docs {
		if x == d {
			return i
		}
	}
	return 0
}

func explorerWidth(total int) int {
	w := min(max(total*3/10, 24), 40)
	if total-w < 30 {
		return 0
	}
	return w
}

func (e *Editor) explorerPanel(rows, width int) []render.StyledLine {
	x := &e.explorer
	inner := width - 1
	focused := e.mode == ModeExplorer
	border := render.Span{Text: "│", Style: render.Style{Fg: colFgGutter}}

	visible := rows - 1
	if x.sel < x.top {
		x.top = x.sel
	}
	if x.sel >= x.top+visible {
		x.top = x.sel - visible + 1
	}
	x.top = max(min(x.top, len(x.rows)-visible), 0)

	out := make([]render.StyledLine, 0, rows)
	header := render.StyledLine{{Text: " " + sanitize(x.tree.name), Style: render.Style{Fg: colBlue, Bold: true}}}
	out = append(out, append(padTo(truncateLine(header, inner), inner, render.Style{}), border))

	for i := x.top; i < x.top+visible; i++ {
		if i >= len(x.rows) {
			out = append(out, render.StyledLine{{Text: strings.Repeat(" ", inner)}, border})
			continue
		}
		n := x.rows[i]
		icon := "  "
		style := render.Style{Fg: colFg}
		switch {
		case n.dir && n.expanded:
			icon, style = "▾ ", render.Style{Fg: colBlue}
		case n.dir:
			icon, style = "▸ ", render.Style{Fg: colBlue}
		case strings.HasSuffix(n.name, ".go"):
			style = render.Style{Fg: colBlue1}
		case strings.HasPrefix(n.name, "."):
			style = render.Style{Fg: colDark3}
		}
		if e.path != "" && samePath(n.path, e.path) {
			style = render.Style{Fg: colOrange, Bold: true}
		}
		name := sanitize(n.name)
		if n.dir {
			name += "/"
		}
		if n.dir && n.loaded && len(n.children) == 0 && n.expanded {
			name += " (vazia)"
		}
		bg := render.Color{}
		if i == x.sel {
			bg = colBgHighlight
			if focused {
				bg = colBgSel
			}
		}
		line := render.StyledLine{
			{Text: strings.Repeat("  ", n.depth) + " ", Style: render.Style{Bg: bg}},
			{Text: icon, Style: render.Style{Fg: colFgGutter, Bg: bg}},
			{Text: name, Style: withBg(style, bg)},
		}
		if line.Width() > inner {
			line = append(truncateLine(line, inner-1), render.Span{Text: "…", Style: render.Style{Fg: colDark3, Bg: bg}})
		}
		out = append(out, append(padTo(line, inner, render.Style{Bg: bg}), border))
	}
	return out
}

func (e *Editor) ShowExplorer() {
	e.openExplorer(true)
}
