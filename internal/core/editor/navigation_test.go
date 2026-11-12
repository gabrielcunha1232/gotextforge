package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func makeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if strings.HasSuffix(rel, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func projectEditor(t *testing.T, files map[string]string) (*Editor, string) {
	t.Helper()
	root := makeTree(t, files)
	ed := New(nil, "")
	ed.save = (&fakeSaver{}).save
	ed.SetRoot(root)
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	return ed, root
}

func rowNames(ed *Editor) []string {
	var out []string
	for _, r := range ed.explorer.rows {
		name := strings.Repeat("  ", r.depth) + r.name
		if r.dir {
			name += "/"
		}
		out = append(out, name)
	}
	return out
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var sampleProject = map[string]string{
	"go.mod":          "module x\n",
	"main.go":         "package main\n",
	"cmd/app/app.go":  "package app\n",
	"internal/a.go":   "package internal\n",
	"internal/b.go":   "package internal\n",
	"docs/":           "",
	".git/config":     "x",
	".env":            "SEGREDO=1",
	"README.md":       "# x\n",
	"internal/vazia/": "",
}

func TestExplorerNavega(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)

	typeKeys(t, ed, " e")
	if ed.Mode() != ModeExplorer || !ed.explorer.open {
		t.Fatalf("Espaco e deveria abrir e focar o explorador (modo %v)", ed.Mode())
	}
	if got := strings.Join(rowNames(ed), ","); got != "cmd/,docs/,internal/,go.mod,main.go,README.md" {
		t.Errorf("raiz = %s (pastas primeiro, ocultos escondidos)", got)
	}

	typeKeys(t, ed, "jjl")
	if got := strings.Join(rowNames(ed), ","); !strings.Contains(got, "internal/,  vazia/,  a.go,  b.go") {
		t.Errorf("l deveria expandir internal/: %s", got)
	}
	typeKeys(t, ed, "l")
	if ed.selectedNode().name != "vazia" {
		t.Errorf("l numa pasta ja aberta desce para o primeiro filho, esta em %q", ed.selectedNode().name)
	}
	typeKeys(t, ed, "h")
	if ed.selectedNode().name != "internal" {
		t.Errorf("h num filho fechado volta para a pasta pai, esta em %q", ed.selectedNode().name)
	}
	typeKeys(t, ed, "h")
	if strings.Contains(strings.Join(rowNames(ed), ","), "a.go") {
		t.Errorf("h numa pasta aberta deveria fechar: %v", rowNames(ed))
	}

	typeKeys(t, ed, "lGk<CR>")
	if ed.Mode() != ModeNormal || ed.filename != "main.go" || bufferText(ed) != "package main" {
		t.Fatalf("Enter num arquivo deveria abrir e focar o editor: modo %v arquivo %q", ed.Mode(), ed.filename)
	}
	if !ed.explorer.open {
		t.Error("abrir um arquivo nao fecha o explorador")
	}

	typeKeys(t, ed, " e")
	if ed.Mode() != ModeExplorer || ed.selectedNode().name != "main.go" {
		t.Errorf("reabrir o foco deveria selecionar o arquivo atual: %v", ed.selectedNode())
	}
	typeKeys(t, ed, " e")
	if ed.explorer.open || ed.Mode() != ModeNormal {
		t.Error("Espaco e com o explorador focado deveria fechar")
	}

	typeKeys(t, ed, " eH")
	if !strings.Contains(strings.Join(rowNames(ed), ","), ".env") {
		t.Errorf("H deveria mostrar os ocultos: %v", rowNames(ed))
	}
	typeKeys(t, ed, "<Esc>")
	if ed.Mode() != ModeNormal || !ed.explorer.open {
		t.Error("Esc deveria voltar ao editor sem fechar o explorador")
	}
	_ = root
}

func TestExplorerTrocaDePasta(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	typeKeys(t, ed, " e")

	typeKeys(t, ed, "<BS>")
	if !samePath(ed.projectRoot(), filepath.Dir(root)) {
		t.Fatalf("Backspace deveria subir uma pasta: raiz %q", ed.projectRoot())
	}
	if sel := ed.selectedNode(); sel == nil || !samePath(sel.path, root) {
		t.Errorf("depois de subir, a pasta de onde viemos fica selecionada: %v", sel)
	}

	typeKeys(t, ed, ".")
	if !samePath(ed.projectRoot(), root) {
		t.Fatalf(". deveria transformar a pasta selecionada na raiz: %q", ed.projectRoot())
	}
	typeKeys(t, ed, "j.")
	if !samePath(ed.projectRoot(), filepath.Join(root, "docs")) {
		t.Errorf(". em docs/ deveria entrar nela: %q", ed.projectRoot())
	}
}

func TestExplorerCriaRenomeiaApaga(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	typeKeys(t, ed, " ejj")

	typeKeys(t, ed, "anovo.go<CR>")
	if !exists(filepath.Join(root, "internal", "novo.go")) {
		t.Fatal("a deveria criar o arquivo dentro da pasta selecionada")
	}
	if ed.filename != "novo.go" || ed.Mode() != ModeNormal {
		t.Errorf("o arquivo criado deveria abrir no editor: %q modo %v", ed.filename, ed.Mode())
	}

	typeKeys(t, ed, " ea")
	typeKeys(t, ed, "pacote/sub/<CR>")
	if fi, err := os.Stat(filepath.Join(root, "internal", "pacote", "sub")); err != nil || !fi.IsDir() {
		t.Errorf("nome terminando com / deveria criar pastas: %v", err)
	}
	typeKeys(t, ed, "a")
	typeKeys(t, ed, "x/y/z.go<CR>")
	if !exists(filepath.Join(root, "internal", "pacote", "sub", "x", "y", "z.go")) {
		t.Errorf("caminho com pastas intermediarias deveria criar tudo")
	}

	typeKeys(t, ed, " e")
	typeKeys(t, ed, "a")
	typeKeys(t, ed, "<CR>")
	if ed.Mode() != ModeExplorer {
		t.Errorf("nome vazio nao cria nada e volta ao explorador")
	}

	ed.OpenFile(filepath.Join(root, "main.go"))
	typeKeys(t, ed, " e")
	if ed.selectedNode().name != "main.go" {
		t.Fatalf("explorador deveria revelar main.go, esta em %q", ed.selectedNode().name)
	}
	typeKeys(t, ed, "r")
	typeKeys(t, ed, "<BS><BS><BS><BS><BS><BS><BS>")
	typeKeys(t, ed, "inicio.go<CR>")
	if exists(filepath.Join(root, "main.go")) || !exists(filepath.Join(root, "inicio.go")) {
		t.Fatal("r deveria renomear main.go para inicio.go")
	}
	found := false
	for _, d := range ed.docs {
		if d.filename == "inicio.go" {
			found = true
		}
	}
	if !found {
		t.Error("a aba aberta deveria acompanhar o novo nome")
	}

	typeKeys(t, ed, "dn<CR>")
	if !exists(filepath.Join(root, "inicio.go")) {
		t.Error("responder n nao deveria apagar")
	}
	typeKeys(t, ed, "ds<CR>")
	if exists(filepath.Join(root, "inicio.go")) {
		t.Error("responder s deveria apagar")
	}
	for _, d := range ed.docs {
		if d.filename == "inicio.go" {
			t.Error("a aba do arquivo apagado deveria fechar")
		}
	}

	if ed.Mode() != ModeExplorer {
		t.Fatalf("depois de apagar, o foco continua no explorador (modo %v)", ed.Mode())
	}
	typeKeys(t, ed, "g")
	for steps := 0; ed.selectedNode().name != "internal"; steps++ {
		if steps > 50 {
			t.Fatalf("internal/ nao aparece no explorador: %v", rowNames(ed))
		}
		typeKeys(t, ed, "j")
	}
	typeKeys(t, ed, "ds<CR>")
	if !exists(filepath.Join(root, "internal")) || !strings.Contains(ed.Frame().Status.Text(), "vazias") {
		t.Errorf("pasta com conteudo nao pode ser apagada: %q", ed.Frame().Status.Text())
	}
}

func TestExplorerNaoApagaArquivoModificado(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	ed.OpenFile(filepath.Join(root, "README.md"))
	typeKeys(t, ed, "x")
	typeKeys(t, ed, " eds<CR>")
	if !exists(filepath.Join(root, "README.md")) || !strings.Contains(ed.Frame().Status.Text(), "não salvas") {
		t.Errorf("nao deveria apagar arquivo com alteracoes nao salvas: %q", ed.Frame().Status.Text())
	}
}

func TestFinder(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	typeKeys(t, ed, "  ")
	if ed.Mode() != ModeFinder || ed.finder == nil {
		t.Fatalf("Espaco Espaco deveria abrir a busca (modo %v)", ed.Mode())
	}
	for _, it := range ed.finder.items {
		if strings.HasPrefix(it, ".git") || it == ".env" {
			t.Errorf("busca nao deveria listar %q", it)
		}
	}
	if len(ed.finder.items) != 6 {
		t.Errorf("deveria listar 6 arquivos, listou %d: %v", len(ed.finder.items), ed.finder.items)
	}

	typeKeys(t, ed, "appgo")
	if len(ed.finder.matches) == 0 || ed.finder.matches[0].Text != "cmd/app/app.go" {
		t.Fatalf("appgo deveria achar cmd/app/app.go: %+v", ed.finder.matches)
	}
	f := ed.Frame()
	all := ""
	for _, l := range f.Lines {
		all += l.Text() + "\n"
	}
	if !strings.Contains(all, "Arquivos") || !strings.Contains(all, "> appgo") || !strings.Contains(all, "▶ cmd/app/app.go") {
		t.Errorf("janela de busca nao desenhada como esperado:\n%s", all)
	}
	if f.Cursor.Col < 4 {
		t.Errorf("cursor deveria estar no campo de busca: %+v", f.Cursor)
	}

	typeKeys(t, ed, "<CR>")
	if ed.Mode() != ModeNormal || ed.filename != "app.go" {
		t.Fatalf("Enter deveria abrir o arquivo: %q", ed.filename)
	}

	typeKeys(t, ed, "  inta")
	pressCtrl(ed, 'n')
	typeKeys(t, ed, "<CR>")
	if ed.filename != "b.go" {
		t.Errorf("Ctrl+n deveria ir para o segundo resultado (b.go), abriu %q", ed.filename)
	}

	typeKeys(t, ed, "  pkg/util/str.go")
	if !ed.finder.createRow() {
		t.Fatal("sem correspondencia, a busca deveria oferecer criar")
	}
	typeKeys(t, ed, "<CR>")
	if !exists(filepath.Join(root, "pkg", "util", "str.go")) || ed.filename != "str.go" {
		t.Errorf("Enter na linha de criar deveria criar e abrir o arquivo")
	}

	typeKeys(t, ed, "  xyz<Esc>")
	if ed.Mode() != ModeNormal || ed.finder != nil {
		t.Error("Esc deveria fechar a busca")
	}

	typeKeys(t, ed, " ,")
	if ed.Mode() != ModeFinder || len(ed.finder.items) != 3 || ed.finder.allowCreate {
		t.Errorf("Espaco , deveria listar as 3 abas abertas: %v", ed.finder.items)
	}
}

func TestNovoArquivo(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	ed.OpenFile(filepath.Join(root, "internal", "a.go"))
	typeKeys(t, ed, " fn")
	if ed.Mode() != ModePrompt || !strings.HasSuffix(ed.Frame().Status.Text(), "novo arquivo: internal/") {
		t.Fatalf("Espaco f n deveria perguntar o nome ja sugerindo a pasta atual: %q", ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "c.go<CR>")
	if !exists(filepath.Join(root, "internal", "c.go")) || ed.filename != "c.go" {
		t.Errorf("deveria criar e abrir internal/c.go")
	}
}

func TestTelaInicialComAcoes(t *testing.T) {
	ed, root := projectEditor(t, sampleProject)
	all := ""
	for _, l := range ed.Frame().Lines {
		all += l.Text() + "\n"
	}
	for _, want := range []string{"buscar arquivo", "explorar a pasta", "novo arquivo", "trocar de pasta", filepath.Base(root)} {
		if !strings.Contains(all, want) {
			t.Errorf("tela inicial deveria mostrar %q", want)
		}
	}

	typeKeys(t, ed, "f")
	if ed.Mode() != ModeFinder {
		t.Error("f na tela inicial abre a busca")
	}
	typeKeys(t, ed, "<Esc>e")
	if ed.Mode() != ModeExplorer {
		t.Error("e na tela inicial abre o explorador")
	}
	typeKeys(t, ed, "<Esc>p")
	if ed.Mode() != ModeExplorer || !strings.Contains(ed.Frame().Status.Text(), "Backspace") {
		t.Errorf("p deveria abrir o explorador explicando como trocar de pasta: %q", ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "<Esc>n")
	if ed.Mode() != ModePrompt {
		t.Error("n na tela inicial pergunta o nome do novo arquivo")
	}
	typeKeys(t, ed, "<Esc>q")
	if !ed.ShouldQuit() {
		t.Error("q na tela inicial sai")
	}
}

func TestLayoutComExplorador(t *testing.T) {
	for _, width := range []int{20, 40, 60, 100, 200} {
		ed, root := projectEditor(t, sampleProject)
		ed.HandleEvent(input.NewResizeEvent(width, 20))
		ed.OpenFile(filepath.Join(root, "main.go"))
		typeKeys(t, ed, " e")
		f := ed.Frame()
		for i, l := range f.Lines {
			if l.Width() > width {
				t.Errorf("largura %d: linha %d com %d celulas", width, i, l.Width())
			}
		}
		if !f.HideCursor {
			t.Errorf("largura %d: com o foco no explorador o cursor do texto fica escondido", width)
		}
		typeKeys(t, ed, "<Esc>")
		f = ed.Frame()
		panel := explorerWidth(width)
		if f.HideCursor || f.Cursor.Col != panel+ed.gutterWidth() {
			t.Errorf("largura %d: cursor do editor deveria estar depois do painel (%d) e do gutter: %+v", width, panel, f.Cursor)
		}
		if panel > 0 && !strings.Contains(f.Lines[0].Text(), "│") {
			t.Errorf("largura %d: painel deveria ter borda", width)
		}
	}
}

func TestShowExplorerComPastaInicial(t *testing.T) {
	root := makeTree(t, sampleProject)
	ed := New(nil, "")
	ed.SetRoot(filepath.Join(root, "internal"))
	ed.ShowExplorer()
	if ed.Mode() != ModeExplorer || !samePath(ed.explorer.tree.path, filepath.Join(root, "internal")) {
		t.Errorf("ShowExplorer deveria abrir focado na pasta escolhida: %q", ed.explorer.tree.path)
	}
	if got := strings.Join(rowNames(ed), ","); got != "vazia/,a.go,b.go" {
		t.Errorf("conteudo = %s", got)
	}
}

func TestTelaInicialEmTerminalBaixo(t *testing.T) {
	for _, h := range []int{14, 16, 20, 30} {
		ed, _ := projectEditor(t, sampleProject)
		ed.HandleEvent(input.NewResizeEvent(100, h))
		all := ""
		for _, l := range ed.Frame().Lines {
			all += l.Text() + "\n"
		}
		if !strings.Contains(all, "buscar arquivo") {
			t.Errorf("altura %d: as acoes da tela inicial deveriam aparecer mesmo sem espaco para o logo", h)
		}
	}
}

func TestExplorerNomeCortadoTemReticencias(t *testing.T) {
	ed, _ := projectEditor(t, map[string]string{"um_nome_de_arquivo_bem_comprido_mesmo.go": ""})
	ed.HandleEvent(input.NewResizeEvent(80, 20))
	typeKeys(t, ed, " e")
	row := ed.Frame().Lines[1].Text()
	if !strings.Contains(row, "…│") {
		t.Errorf("nome cortado deveria terminar com …: %q", row)
	}
}
