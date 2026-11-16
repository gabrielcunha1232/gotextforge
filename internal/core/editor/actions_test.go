package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
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

func TestTerminalClickEKinds(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n\nfunc f() {\n\tx := 1\n}\n"), 0o644)
	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	ed.SetRoot(dir)
	ed.term.open = true
	ed.term.job = &runner.Job{Command: "go build ./..."}
	ed.term.snap = runner.Snapshot{Total: 4, State: runner.Failed, Lines: []string{
		"# exemplo",
		"./a.go:4:2: declared and not used: x",
		"ok  \tpkg\t0.1s",
		"FAIL",
	}}

	if k := ed.UI().Terminal.Kinds; k != " loe" {
		t.Errorf("kinds = %q, quero %q", k, " loe")
	}

	ed.TerminalClick(0)
	if ed.mode != ModeTerminal || ed.term.sel != 0 || ed.term.follow {
		t.Errorf("clicar numa linha comum deveria só selecionar e focar: modo %v sel %d", ed.mode, ed.term.sel)
	}
	ed.TerminalClick(1)
	if ed.mode != ModeNormal || filepath.Base(ed.path) != "a.go" || ed.cursor.Line != 3 || ed.cursor.Col != 1 {
		t.Errorf("clicar no erro deveria abrir a.go na linha 4 coluna 2: %q %+v modo %v", ed.path, ed.cursor, ed.mode)
	}

	ed.TerminalAction("close")
	if ed.term.open {
		t.Errorf("close deveria fechar o terminal")
	}
}

func TestGoTo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "b.go")
	os.WriteFile(path, []byte("package b\n\nvar s = \"x\"\n"), 0o644)
	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	typeKeys(t, ed, "i")

	ed.GoTo(path, 3, 5)
	if ed.mode != ModeNormal || ed.cursor.Line != 2 || ed.cursor.Col != 4 {
		t.Errorf("GoTo deveria sair do insert e ir para 3:5: modo %v cursor %+v", ed.mode, ed.cursor)
	}
	ed.GoTo(path, 99, 99)
	if ed.cursor.Line != 2 {
		t.Errorf("linha fora do arquivo deveria parar na última: %+v", ed.cursor)
	}
}

func TestDockerCliques(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n"}, "a.go")
	ed.dockerExec = fakeDockerExec(nil)
	ed.TogglePanel("docker")
	ed.receiveDocker(dockerSnapshot{gen: ed.dock.gen, containers: []docker.Container{{ID: "1", Name: "api", State: "running"}, {ID: "2", Name: "db", State: "exited"}}})
	defer ed.Close()

	ed.DockerClick(1)
	if ed.dock.sel != 1 || ed.mode != ModeDocker {
		t.Errorf("DockerClick(1): sel %d modo %v", ed.dock.sel, ed.mode)
	}
	ed.DockerTab("Imagens")
	if ed.dock.tab != dockImages || ed.dock.sel != 0 {
		t.Errorf("DockerTab: aba %v sel %d", ed.dock.tab, ed.dock.sel)
	}
	ed.DockerKey("x")
	if !strings.Contains(ed.Frame().Status.Text(), "não é um terminal") {
		t.Errorf("x deveria avisar que o painel não é terminal: %q", ed.Frame().Status.Text())
	}
	ed.DockerKey("Z")
	ed.DockerKey("q")
	if ed.dock.open {
		t.Errorf("q deveria fechar o painel")
	}
}

func TestFinderNaMolduraExterna(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "alfa.go"), []byte("package x\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "beta.go"), []byte("package x\n"), 0o644)
	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	ed.SetRoot(dir)
	ed.SetExternalChrome(true)

	before := ed.Frame()
	ed.TogglePanel("finder")
	f := ed.Frame()
	if !f.HideCursor {
		t.Errorf("com a busca aberta (desenhada pelo HTML), o cursor do canvas deveria sumir")
	}
	for i := range f.Lines {
		if f.Lines[i].Text() != before.Lines[i].Text() {
			t.Fatalf("na moldura externa a busca não deveria ser desenhada no canvas (linha %d: %q)", i, f.Lines[i].Text())
		}
	}

	typeKeys(t, ed, "bet")
	ed.FinderClick(5)
	if ed.mode != ModeFinder {
		t.Fatalf("índice inválido não deveria fechar a busca")
	}
	ed.FinderClick(0)
	if ed.mode != ModeNormal || filepath.Base(ed.path) != "beta.go" {
		t.Errorf("clicar no resultado deveria abrir beta.go: modo %v, aba %q", ed.mode, ed.path)
	}

	ed.TogglePanel("finder")
	typeKeys(t, ed, "novo.go")
	if ui := ed.UI().Finder; ui == nil || !ui.Create {
		t.Fatalf("sem resultado, deveria oferecer criar: %+v", ui)
	}
	ed.FinderClick(0)
	if _, err := os.Stat(filepath.Join(dir, "novo.go")); err != nil || filepath.Base(ed.path) != "novo.go" {
		t.Errorf("clicar em + criar deveria criar e abrir novo.go: %v, aba %q", err, ed.path)
	}
}

func TestClickText(t *testing.T) {
	src := "package a\n\nfunc f() {\n\tx := 1\n}\n"
	ed, _ := editorWithFiles(t, map[string]string{"a.go": src}, "a.go")
	ed.SetExternalChrome(true)
	gw := ed.gutterWidth()

	ed.ClickText(3, gw+5)
	if ed.cursor.Line != 3 || ed.cursor.Col != 2 {
		t.Errorf("clique depois da tab: cursor %+v, quero linha 3 coluna 2", ed.cursor)
	}
	ed.ClickText(3, gw+2)
	if ed.cursor.Col != 0 {
		t.Errorf("clique no meio da tab deveria cair na tab: %+v", ed.cursor)
	}
	ed.ClickText(0, gw+80)
	if ed.cursor.Line != 0 || ed.cursor.Col != 8 {
		t.Errorf("clique depois do fim da linha (modo normal) para no último caractere: %+v", ed.cursor)
	}
	ed.ClickText(20, 0)
	if ed.cursor.Line != 4 {
		t.Errorf("clique abaixo do texto vai para a última linha: %+v", ed.cursor)
	}

	typeKeys(t, ed, "ggi")
	ed.ClickText(0, gw+80)
	if ed.mode != ModeInsert || ed.cursor.Col != 9 {
		t.Errorf("no insert, o clique mantém o modo e pode ir até depois do último caractere: modo %v %+v", ed.mode, ed.cursor)
	}
	typeKeys(t, ed, "<Esc>:")
	ed.ClickText(1, gw)
	if ed.mode != ModeNormal || ed.cursor.Line != 1 {
		t.Errorf("clique com : aberto cancela o comando e posiciona: modo %v %+v", ed.mode, ed.cursor)
	}
}

func TestScrollText(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("linha\n")
	}
	ed, _ := editorWithFiles(t, map[string]string{"a.txt": sb.String()}, "a.txt")
	ed.SetExternalChrome(true)

	ed.ScrollText(10)
	if ed.viewport.Top != 10 || ed.cursor.Line != 10 {
		t.Errorf("rolar 10 deveria levar o topo e o cursor para a linha 10: topo %d cursor %d", ed.viewport.Top, ed.cursor.Line)
	}
	ed.ScrollText(-5)
	if ed.viewport.Top != 5 || ed.cursor.Line != 10 {
		t.Errorf("rolar para cima com o cursor visível não deveria mexer nele: topo %d cursor %d", ed.viewport.Top, ed.cursor.Line)
	}
	ed.ScrollText(1000)
	if ed.cursor.Line < ed.viewport.Top {
		t.Errorf("cursor fora da tela depois de rolar até o fim: topo %d cursor %d", ed.viewport.Top, ed.cursor.Line)
	}
}

func TestRequestQuit(t *testing.T) {
	ed, saver := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.RequestQuit()
	if !ed.ShouldQuit() {
		t.Fatal("sem alterações, deveria sair direto")
	}

	ed, saver = editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	typeKeys(t, ed, "Ax<Esc>")
	ed.SwitchTab(0)
	typeKeys(t, ed, "Ay<Esc>")
	ed.RequestQuit()
	if ed.ShouldQuit() || !strings.Contains(ed.Frame().Status.Text(), "2 arquivos têm alterações") {
		t.Fatalf("deveria perguntar: %q", ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "c<CR>")
	if ed.ShouldQuit() {
		t.Fatal("cancelar não deveria sair")
	}
	ed.RequestQuit()
	typeKeys(t, ed, "s<CR>")
	if !ed.ShouldQuit() || saver.calls != 2 {
		t.Errorf("sim deveria salvar os 2 e sair: quit %v, salvos %d", ed.ShouldQuit(), saver.calls)
	}
}

func TestOpenFolder(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(a, "a.go"), []byte("package a\n"), 0o644)
	os.WriteFile(filepath.Join(b, "b.go"), []byte("package b\n"), 0o644)

	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	ed.SetRoot(a)
	ed.OpenExplorer()
	ed.term.cwd = filepath.Join(a, "sub")
	typeKeys(t, ed, "i")

	ed.OpenFolder(b)
	if !samePath(ed.Root(), b) || ed.mode != ModeNormal {
		t.Fatalf("raiz %q modo %v, quero %q em NORMAL", ed.Root(), ed.mode, b)
	}
	ui := ed.UI()
	if len(ui.Explorer.Rows) != 1 || ui.Explorer.Rows[0].Name != "b.go" {
		t.Errorf("a árvore deveria mostrar a pasta nova: %+v", ui.Explorer.Rows)
	}
	if ed.terminalDir() != ed.Root() {
		t.Errorf("o terminal deveria seguir a pasta nova: %q", ed.terminalDir())
	}

	ed.OpenFolder(filepath.Join(b, "b.go"))
	if !samePath(ed.Root(), b) || !strings.Contains(ed.message, "não encontrada") {
		t.Errorf("um arquivo não é pasta: raiz %q, mensagem %q", ed.Root(), ed.message)
	}
}
