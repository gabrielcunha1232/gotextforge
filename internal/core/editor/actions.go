package editor

import (
	"fmt"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

func (e *Editor) escape() {
	e.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyEsc}))
}

func (e *Editor) SwitchTab(i int) {
	e.escape()
	if i < 0 || i >= len(e.docs) {
		return
	}
	e.switchTo(i)
	e.viewport.EnsureVisible(e.cursor.Line, e.buf.LineCount())
	e.syncLanguageServer()
}

func (e *Editor) CloseTab(i int) {
	e.escape()
	if i < 0 || i >= len(e.docs) {
		return
	}
	e.switchTo(i)
	if !e.buf.Modified() {
		e.closeTab()
		return
	}
	label := fmt.Sprintf("%s tem alterações. Salvar? (s)im, (n)ão, (c)ancelar: ", e.filename)
	e.openPrompt(label, "", func(answer string) {
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "s", "sim":
			if e.write() {
				e.closeTab()
			}
		case "n", "não", "nao":
			e.closeTab()
		default:
			e.message = "fechar cancelado"
		}
	})
}

func (e *Editor) closeTab() {
	if len(e.docs) > 1 {
		e.closeCurrent(true)
		return
	}
	if e.lsp.client != nil && e.opened {
		e.lsp.client.CloseDocument(e.path)
	}
	doc := newDocument(nil, "")
	e.docs = []*document{doc}
	e.document = doc
	e.layout()
}

func (e *Editor) OpenExplorer() {
	e.openExplorer(false)
}

func (e *Editor) ExplorerClick(path string) {
	e.escape()
	e.openExplorer(false)
	if !e.selectPath(path) {
		return
	}
	n := e.selectedNode()
	e.explorerOpenNode(n, true)
	if n.dir {
		e.mode = ModeExplorer
	}
}

func (e *Editor) TogglePanel(name string) {
	was := e.mode
	e.escape()
	switch name {
	case "explorer":
		if e.explorer.open && was == ModeExplorer {
			e.explorer.open = false
		} else {
			e.openExplorer(true)
		}
	case "terminal":
		if e.term.open && was == ModeTerminal {
			e.closeTerminal()
		} else {
			e.openTerminal(true)
		}
	case "docker":
		switch {
		case !e.dock.open:
			e.openDocker()
		case was == ModeDocker:
			e.closeDocker()
		default:
			e.mode = ModeDocker
		}
	case "finder":
		if was != ModeFinder {
			e.openFileFinder()
		}
	}
}

func (e *Editor) GoTo(path string, line, col int) {
	e.escape()
	e.goTo(path, line, col)
}

func (e *Editor) TerminalClick(i int) {
	e.escape()
	t := &e.term
	if !t.open || i < 0 || i >= len(t.snap.Lines) {
		return
	}
	t.sel, t.follow = i, false
	e.mode = ModeTerminal
	if _, _, _, ok := parseLocation(t.snap.Lines[i]); ok {
		e.jumpToLocation(i)
	}
}

func (e *Editor) TerminalAction(name string) {
	e.escape()
	t := &e.term
	switch name {
	case "rerun":
		e.rerunCommand()
	case "cancel":
		if t.job != nil && t.snap.State == runner.Running {
			t.job.Cancel()
		}
	case "new":
		e.promptCommand()
	case "close":
		e.closeTerminal()
	}
}

func (e *Editor) DockerTab(name string) {
	e.escape()
	for i, n := range dockerTabNames {
		if n == name && e.dock.open {
			e.dock.tab = dockerTab(i)
			e.dock.sel, e.dock.top = 0, 0
			e.mode = ModeDocker
		}
	}
}

func (e *Editor) DockerClick(i int) {
	e.escape()
	if !e.dock.open || i < 0 || i >= e.dockerItemCount() {
		return
	}
	e.dock.sel = i
	e.mode = ModeDocker
}

func (e *Editor) DockerKey(key string) {
	e.escape()
	if !e.dock.open || len(key) != 1 || !strings.Contains("slrbdxq", key) {
		return
	}
	e.mode = ModeDocker
	e.handleDockerKey(input.Key{Code: input.KeyRune, Rune: rune(key[0])})
}

func (e *Editor) FinderClick(i int) {
	f := e.finder
	if e.mode != ModeFinder || f == nil || i < 0 || i >= f.rowCount() {
		return
	}
	f.sel = i
	e.handleFinderKey(input.Key{Code: input.KeyEnter})
}
