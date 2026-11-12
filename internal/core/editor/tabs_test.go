package editor

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
)

func fakeFiles(files map[string]string) LoadFunc {
	return func(path string) (*buffer.Buffer, error) {
		src, ok := files[path]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return buffer.Load(strings.NewReader(src))
	}
}

func editorWithFiles(t *testing.T, files map[string]string, open ...string) (*Editor, *fakeSaver) {
	t.Helper()
	ed := New(nil, "")
	ed.load = fakeFiles(files)
	fs := &fakeSaver{}
	ed.save = fs.save
	ed.HandleEvent(input.NewResizeEvent(80, 24))
	for _, p := range open {
		ed.OpenFile(p)
	}
	return ed, fs
}

func TestAbrirArquivos(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go")
	if len(ed.docs) != 1 || ed.path != "a.go" {
		t.Fatalf("o primeiro arquivo deveria substituir a tela inicial: %d docs, atual %q", len(ed.docs), ed.path)
	}

	typeKeys(t, ed, ":e b.go<CR>")
	if len(ed.docs) != 2 || ed.path != "b.go" || bufferText(ed) != "package b" {
		t.Fatalf(":e deveria abrir b.go numa aba nova: %d docs, atual %q", len(ed.docs), ed.path)
	}

	typeKeys(t, ed, ":e a.go<CR>")
	if len(ed.docs) != 2 || ed.path != "a.go" {
		t.Errorf("abrir de novo um arquivo ja aberto deveria so trocar de aba: %d docs, atual %q", len(ed.docs), ed.path)
	}

	typeKeys(t, ed, ":e novo.go<CR>")
	if ed.path != "novo.go" || !strings.Contains(ed.Frame().Status.Text(), "[Novo]") {
		t.Errorf("arquivo inexistente deveria abrir como novo: %q / %q", ed.path, ed.Frame().Status.Text())
	}

	typeKeys(t, ed, ":e<CR>")
	if !strings.Contains(ed.Frame().Status.Text(), "E32") {
		t.Errorf(":e sem arquivo deveria avisar: %q", ed.Frame().Status.Text())
	}
}

func TestTrocarDeAba(t *testing.T) {
	files := map[string]string{"a.go": "a1\na2\na3\n", "b.go": "b1\nb2\n", "c.go": "c1\n"}
	ed, _ := editorWithFiles(t, files, "a.go", "b.go", "c.go")
	if ed.path != "c.go" {
		t.Fatalf("depois de abrir, a aba atual deveria ser a ultima: %q", ed.path)
	}

	steps := []struct {
		keys string
		want string
	}{
		{"H", "b.go"}, {"H", "a.go"}, {"H", "c.go"}, {"L", "a.go"},
		{"gt", "b.go"}, {"gT", "a.go"}, {"2L", "c.go"},
		{":bn<CR>", "a.go"}, {":bp<CR>", "c.go"},
	}
	for _, s := range steps {
		typeKeys(t, ed, s.keys)
		if ed.path != s.want {
			t.Fatalf("depois de %q a aba atual e %q, esperava %q", s.keys, ed.path, s.want)
		}
	}

	typeKeys(t, ed, "Lj")
	typeKeys(t, ed, "LLL")
	if ed.path != "a.go" || ed.Cursor() != pos(1, 0) {
		t.Errorf("cada aba deveria lembrar o proprio cursor: %q %+v", ed.path, ed.Cursor())
	}
}

func TestFecharAbas(t *testing.T) {
	files := map[string]string{"a.go": "a\n", "b.go": "b\n"}
	ed, fs := editorWithFiles(t, files, "a.go", "b.go")

	typeKeys(t, ed, "x:q<CR>")
	if len(ed.docs) != 2 || !strings.Contains(ed.Frame().Status.Text(), "E37") {
		t.Fatalf(":q com alteracoes deveria avisar e nao fechar")
	}
	typeKeys(t, ed, ":qa<CR>")
	if ed.ShouldQuit() || !strings.Contains(ed.Frame().Status.Text(), "b.go") {
		t.Fatalf(":qa com alteracoes deveria avisar qual arquivo: %q", ed.Frame().Status.Text())
	}

	typeKeys(t, ed, ":wq<CR>")
	if fs.calls != 1 || len(ed.docs) != 1 || ed.path != "a.go" || ed.ShouldQuit() {
		t.Fatalf(":wq deveria salvar e fechar so a aba atual: saves=%d docs=%d atual=%q", fs.calls, len(ed.docs), ed.path)
	}
	typeKeys(t, ed, ":q<CR>")
	if !ed.ShouldQuit() {
		t.Error(":q na ultima aba deveria sair do editor")
	}
}

func TestEscreverTodas(t *testing.T) {
	files := map[string]string{"a.go": "a\n", "b.go": "b\n", "c.go": "c\n"}
	ed, fs := editorWithFiles(t, files, "a.go", "b.go", "c.go")
	typeKeys(t, ed, "xHHx")
	typeKeys(t, ed, ":wa<CR>")
	if fs.calls != 2 {
		t.Errorf(":wa deveria salvar so os 2 arquivos modificados, salvou %d", fs.calls)
	}
	if ed.firstModified() != nil || ed.path != "a.go" {
		t.Errorf("depois de :wa nada fica modificado e a aba atual nao muda (%q)", ed.path)
	}
	typeKeys(t, ed, ":qa<CR>")
	if !ed.ShouldQuit() {
		t.Error(":qa depois de :wa deveria sair")
	}
}

func TestLinhaDeAbas(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "a\n", "b.go": "b\n"}, "a.go")
	if f := ed.Frame(); strings.Contains(f.Lines[0].Text(), "a.go") {
		t.Errorf("com um arquivo so, nao deveria haver linha de abas: %q", f.Lines[0].Text())
	}

	typeKeys(t, ed, ":e b.go<CR>x")
	f := ed.Frame()
	tabs := f.Lines[0].Text()
	if !strings.Contains(tabs, "1 a.go") || !strings.Contains(tabs, "2 b.go ●") {
		t.Errorf("linha de abas = %q", tabs)
	}
	if f.Lines[0].Width() != 80 {
		t.Errorf("linha de abas deveria ocupar a largura toda: %d", f.Lines[0].Width())
	}
	if f.Cursor.Row != 1 {
		t.Errorf("com a linha de abas, o cursor da linha 1 fica na linha 1 da tela: %+v", f.Cursor)
	}
	if len(f.Lines) != 23 {
		t.Errorf("tela 24 = abas + 22 de texto + status; Lines = %d", len(f.Lines))
	}
	if ed.viewport.Height != 22 {
		t.Errorf("viewport deveria encolher para 22, esta %d", ed.viewport.Height)
	}

	typeKeys(t, ed, ":q!<CR>")
	if ed.viewport.Height != 23 || strings.Contains(ed.Frame().Lines[0].Text(), "b.go") {
		t.Errorf("fechando a segunda aba, a linha de abas some e o viewport volta a 23 (%d)", ed.viewport.Height)
	}
}

func TestLinhaDeAbasMostraAAtualEmTelaEstreita(t *testing.T) {
	files := map[string]string{}
	var open []string
	for _, n := range []string{"um.go", "dois.go", "tres.go", "quatro.go", "cinco.go", "seis.go"} {
		files[n] = "x\n"
		open = append(open, n)
	}
	ed, _ := editorWithFiles(t, files, open...)
	ed.HandleEvent(input.NewResizeEvent(30, 10))
	tabs := ed.Frame().Lines[0]
	if !strings.Contains(tabs.Text(), "seis.go") || tabs.Width() != 30 {
		t.Errorf("a aba atual deveria aparecer mesmo com pouco espaco: %q (%d)", tabs.Text(), tabs.Width())
	}
}

func TestAbasComLSP(t *testing.T) {
	files := map[string]string{"a.go": "package a\n", "b.go": "package b\n", "leia.md": "# oi\n"}
	ed, _ := editorWithFiles(t, files)
	fc := &fakeClient{events: make(chan lsp.Event, 4)}
	starts := 0
	ed.SetLanguageServer(func(string) LanguageClient { starts++; return fc })

	ed.OpenFile("leia.md")
	if starts != 0 {
		t.Fatal("arquivo que nao e .go nao deveria iniciar o gopls")
	}
	ed.OpenFile("a.go")
	ed.OpenFile("b.go")
	if starts != 1 {
		t.Errorf("um gopls so para todos os arquivos; iniciou %d", starts)
	}
	if len(fc.paths) != 2 || fc.paths[0] != "a.go" || fc.paths[1] != "b.go" {
		t.Errorf("didOpen deveria ir para a.go e b.go: %v", fc.paths)
	}

	ed.HandleAsync(lsp.StatusEvent{Ready: true})
	ed.HandleAsync(lsp.DiagnosticsEvent{URI: lsp.PathToURI("a.go"), Items: []lsp.Diagnostic{
		{Range: rng(0, 0, 0, 7), Severity: lsp.SeverityError, Message: "erro em a"},
	}})
	if errs, _ := ed.diagnosticCounts(); errs != 0 {
		t.Errorf("a aba atual (b.go) nao deveria mostrar os erros de a.go")
	}
	typeKeys(t, ed, "H")
	if errs, _ := ed.diagnosticCounts(); errs != 1 {
		t.Errorf("em a.go os erros deveriam aparecer")
	}

	typeKeys(t, ed, "Ax<Esc>")
	if len(fc.changes) != 1 {
		t.Errorf("editar a.go deveria mandar Change: %v", fc.changes)
	}
	typeKeys(t, ed, "u:q<CR>")
	if len(fc.closedDocs) != 1 || fc.closedDocs[0] != "a.go" {
		t.Errorf("fechar a aba deveria mandar didClose: %v", fc.closedDocs)
	}
}

func TestIntroSomeComMaisDeUmaAba(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "x\n"})
	if !ed.showIntro() {
		t.Fatal("sem arquivos deveria mostrar a tela inicial")
	}
	typeKeys(t, ed, ":e a.go<CR>")
	if ed.showIntro() || len(ed.docs) != 1 {
		t.Errorf("abrir um arquivo a partir da tela inicial substitui a tela (docs=%d)", len(ed.docs))
	}
}
