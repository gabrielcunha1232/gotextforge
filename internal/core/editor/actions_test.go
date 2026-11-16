package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TestSwitchTabSaiDoInsert(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	typeKeys(t, ed, "Ax")
	ed.SwitchTab(0)
	if ed.path != "a.go" || ed.mode != ModeNormal {
		t.Fatalf("aba %q modo %v, quero a.go em NORMAL", ed.path, ed.mode)
	}
	ed.SwitchTab(1)
	typeKeys(t, ed, "u")
	if bufferText(ed) != "package b" {
		t.Errorf("o x digitado antes do clique deveria ser um passo de undo só: %q", bufferText(ed))
	}
	ed.SwitchTab(7)
	if ed.path != "b.go" {
		t.Errorf("índice inválido não deveria trocar de aba")
	}
}

func TestCloseTabSemAlteracoes(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.CloseTab(0)
	if len(ed.docs) != 1 || ed.path != "b.go" {
		t.Fatalf("docs %d, atual %q", len(ed.docs), ed.path)
	}
	ed.CloseTab(0)
	if ed.ShouldQuit() {
		t.Fatal("fechar a última aba pelo clique não deveria sair do editor")
	}
	if len(ed.docs) != 1 || ed.path != "" || !ed.showIntro() {
		t.Errorf("deveria voltar à tela inicial: docs %d, caminho %q", len(ed.docs), ed.path)
	}
}

func TestCloseTabComAlteracoesPergunta(t *testing.T) {
	ed, saver := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.SwitchTab(0)
	typeKeys(t, ed, "Ax<Esc>")
	ed.SwitchTab(1)

	ed.CloseTab(0)
	if ed.mode != ModePrompt || ed.path != "a.go" || !strings.Contains(ed.Frame().Status.Text(), "a.go tem alterações") {
		t.Fatalf("deveria perguntar na aba a.go: modo %v, aba %q, linha %q", ed.mode, ed.path, ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "c<CR>")
	if len(ed.docs) != 2 || !ed.buf.Modified() {
		t.Fatalf("cancelar não deveria fechar nem salvar")
	}

	ed.CloseTab(0)
	typeKeys(t, ed, "n<CR>")
	if len(ed.docs) != 1 || ed.path != "b.go" || saver.calls != 0 {
		t.Fatalf("não deveria fechar sem salvar: docs %d, aba %q, salvos %v", len(ed.docs), ed.path, saver.calls)
	}

	typeKeys(t, ed, "Ay<Esc>")
	ed.CloseTab(0)
	typeKeys(t, ed, "s<CR>")
	if saver.calls != 1 || saver.path != "b.go" || ed.path != "" {
		t.Errorf("sim deveria salvar e fechar: salvos %d (%q), aba %q", saver.calls, saver.path, ed.path)
	}
}

func TestExplorerClick(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "pkg"), 0o755)
	os.WriteFile(filepath.Join(dir, "pkg", "util.go"), []byte("package pkg\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)

	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	ed.SetRoot(dir)
	ed.OpenExplorer()
	if !ed.explorer.open || ed.mode != ModeNormal {
		t.Fatalf("OpenExplorer deveria abrir sem tirar o foco: open %v modo %v", ed.explorer.open, ed.mode)
	}

	pkg := filepath.Join(dir, "pkg")
	ed.ExplorerClick(pkg)
	if ed.mode != ModeExplorer || len(ed.UI().Explorer.Rows) != 3 {
		t.Fatalf("clicar na pasta deveria expandir e focar a árvore: modo %v, linhas %d", ed.mode, len(ed.UI().Explorer.Rows))
	}
	ed.ExplorerClick(pkg)
	if len(ed.UI().Explorer.Rows) != 2 {
		t.Errorf("segundo clique deveria recolher: %d linhas", len(ed.UI().Explorer.Rows))
	}

	ed.ExplorerClick(pkg)
	util := filepath.Join(pkg, "util.go")
	ed.ExplorerClick(util)
	if ed.mode != ModeNormal || !samePath(ed.path, util) {
		t.Fatalf("clicar no arquivo deveria abrir e voltar ao código: modo %v, aba %q", ed.mode, ed.path)
	}
	ui := ed.UI().Explorer
	if r := ui.Rows[ui.Selected]; r.Name != "util.go" || !r.Current {
		t.Errorf("a linha clicada deveria ficar selecionada e marcada como atual: %+v", r)
	}

	ed.ExplorerClick(filepath.Join(dir, "nao-existe.go"))
	if !samePath(ed.path, util) {
		t.Errorf("caminho fora da árvore não deveria fazer nada")
	}
}

func TestTogglePanel(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n"}, "a.go")
	ed.SetRoot(t.TempDir())
	ed.dockerExec = nil

	ed.TogglePanel("explorer")
	if !ed.explorer.open || ed.mode != ModeExplorer {
		t.Fatalf("primeiro clique deveria abrir e focar o explorador")
	}
	ed.TogglePanel("explorer")
	if ed.explorer.open || ed.mode != ModeNormal {
		t.Errorf("clique com o explorador focado deveria fechar")
	}

	ed.TogglePanel("terminal")
	if !ed.term.open || ed.mode != ModeTerminal {
		t.Fatalf("deveria abrir e focar o terminal")
	}
	typeKeys(t, ed, "<Esc>")
	ed.TogglePanel("terminal")
	if !ed.term.open || ed.mode != ModeTerminal {
		t.Errorf("terminal aberto mas sem foco: o clique deveria só focar")
	}
	ed.TogglePanel("terminal")
	if ed.term.open {
		t.Errorf("terminal focado: o clique deveria fechar")
	}

	ed.TogglePanel("finder")
	if ed.mode != ModeFinder {
		t.Fatalf("deveria abrir a busca de arquivos")
	}
	ed.TogglePanel("finder")
	if ed.mode != ModeNormal {
		t.Errorf("segundo clique deveria fechar a busca")
	}

	ed.TogglePanel("docker")
	if ed.dock.open || !strings.Contains(ed.Frame().Status.Text(), "docker desativado") {
		t.Errorf("sem docker, deveria só avisar")
	}
}
