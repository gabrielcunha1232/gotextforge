package editor

import (
	"errors"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type fakeClient struct {
	opened   []string
	changes  []string
	versions []int
	saves    int
	hovers   []buffer.Pos
	comps    []buffer.Pos
	lastID   int
	closed   bool
	events   chan lsp.Event
}

func (f *fakeClient) Open(text string, version int) {
	f.opened = append(f.opened, text)
	f.versions = append(f.versions, version)
}
func (f *fakeClient) Change(text string, version int) {
	f.changes = append(f.changes, text)
	f.versions = append(f.versions, version)
}
func (f *fakeClient) Save() { f.saves++ }
func (f *fakeClient) Hover(id int, _ []rune, p buffer.Pos) {
	f.lastID = id
	f.hovers = append(f.hovers, p)
}
func (f *fakeClient) Completion(id int, _ []rune, p buffer.Pos) {
	f.lastID = id
	f.comps = append(f.comps, p)
}
func (f *fakeClient) Events() <-chan lsp.Event { return f.events }
func (f *fakeClient) Close()                   { f.closed = true }

func editorWithLSP(t *testing.T, src string) (*Editor, *fakeClient) {
	t.Helper()
	ed := editorFrom(t, src)
	ed.path = "/projeto/teste.go"
	fs := &fakeSaver{}
	ed.save = fs.save
	fc := &fakeClient{events: make(chan lsp.Event, 8)}
	ed.SetLanguageClient(fc)
	ed.HandleAsync(lsp.StatusEvent{Ready: true})
	return ed, fc
}

func rng(l1, c1, l2, c2 int) lsp.Range {
	return lsp.Range{Start: lsp.Position{Line: l1, Character: c1}, End: lsp.Position{Line: l2, Character: c2}}
}

func TestSincronizaTextoComOServidor(t *testing.T) {
	ed, fc := editorWithLSP(t, "package main\n")
	if len(fc.opened) != 1 || fc.opened[0] != "package main\n" {
		t.Fatalf("Open = %q", fc.opened)
	}
	typeKeys(t, ed, "jjk")
	if len(fc.changes) != 0 {
		t.Errorf("movimentos nao deveriam mandar Change: %q", fc.changes)
	}
	typeKeys(t, ed, "Ax")
	if len(fc.changes) != 1 || fc.changes[0] != "package mainx\n" {
		t.Errorf("Change = %q", fc.changes)
	}
	for i := 1; i < len(fc.versions); i++ {
		if fc.versions[i] <= fc.versions[i-1] {
			t.Errorf("versoes deveriam crescer: %v", fc.versions)
		}
	}
	typeKeys(t, ed, "<Esc>:w<CR>")
	if fc.saves != 1 {
		t.Errorf("Save chamado %d vezes, esperava 1", fc.saves)
	}
	ed.Close()
	if !fc.closed {
		t.Error("Close do editor deveria fechar o cliente")
	}
}

func TestDiagnosticosNaTela(t *testing.T) {
	ed, _ := editorWithLSP(t, "package main\n\nvar x int = \"texto\"\nvar y = 1\n")
	ed.HandleAsync(lsp.DiagnosticsEvent{Items: []lsp.Diagnostic{
		{Range: rng(2, 12, 2, 19), Severity: lsp.SeverityError, Message: "cannot use \"texto\" as int\nsegunda linha"},
		{Range: rng(3, 4, 3, 5), Severity: lsp.SeverityWarning, Message: "y nao usado"},
	}})
	f := ed.Frame()

	var undercurled strings.Builder
	for _, s := range f.Lines[2] {
		if s.Style.Undercurl {
			undercurled.WriteString(s.Text)
			if s.Style.Sp != colRed {
				t.Errorf("sublinhado do erro deveria ser vermelho, veio %+v", s.Style.Sp)
			}
		}
	}
	if undercurled.String() != "\"texto\"" {
		t.Errorf("trecho sublinhado = %q, esperava %q", undercurled.String(), "\"texto\"")
	}

	text := f.Lines[2].Text()
	if !strings.Contains(text, "● cannot use \"texto\" as int") || strings.Contains(text, "segunda linha") {
		t.Errorf("virtual text da linha 3 = %q", text)
	}
	if f.Lines[2][0].Style.Fg != colRed {
		t.Errorf("numero da linha com erro deveria ser vermelho")
	}
	if f.Lines[3][0].Style.Fg != colYellow {
		t.Errorf("numero da linha com aviso deveria ser amarelo")
	}

	status := f.Status.Text()
	if !strings.Contains(status, "E1") || !strings.Contains(status, "W1") || !strings.Contains(status, "gopls") {
		t.Errorf("status = %q", status)
	}
	for i, l := range f.Lines {
		if l.Width() > 80 {
			t.Errorf("linha %d passou da largura: %d", i, l.Width())
		}
	}
}

func TestDiagnosticoComPosicaoUTF16(t *testing.T) {
	ed, _ := editorWithLSP(t, "s := \"😀\" + x\n")
	ed.HandleAsync(lsp.DiagnosticsEvent{Items: []lsp.Diagnostic{
		{Range: rng(0, 12, 0, 13), Severity: lsp.SeverityError, Message: "x indefinido"},
	}})
	var under string
	for _, s := range ed.Frame().Lines[0] {
		if s.Style.Undercurl {
			under += s.Text
		}
	}
	if under != "x" {
		t.Errorf("com emoji (2 unidades UTF-16) antes, o sublinhado deveria cair no x; caiu em %q", under)
	}
}

func TestDiagnosticoForaDoArquivoNaoQuebra(t *testing.T) {
	ed, _ := editorWithLSP(t, "a\n")
	ed.HandleAsync(lsp.DiagnosticsEvent{Items: []lsp.Diagnostic{
		{Range: rng(50, 0, 60, 3), Severity: lsp.SeverityError, Message: "velho"},
		{Range: rng(0, 40, 0, 99), Severity: lsp.SeverityHint, Message: "alem do fim"},
	}})
	_ = ed.Frame()
}

func TestHover(t *testing.T) {
	ed, fc := editorWithLSP(t, "package main\n\nfunc main() {\n\tprintln(1)\n}\n")
	typeKeys(t, ed, "jjjwK")
	if len(fc.hovers) != 1 || fc.hovers[0] != pos(3, 1) {
		t.Fatalf("hover pedido em %v", fc.hovers)
	}

	ed.HandleAsync(lsp.HoverEvent{ID: fc.lastID, Text: "```go\nfunc println(args ...Type)\n```\n\n---\n\nThe println built-in function [docs](https://x)."})
	f := ed.Frame()
	var all strings.Builder
	for _, l := range f.Lines {
		all.WriteString(l.Text())
		all.WriteByte('\n')
	}
	got := all.String()
	for _, want := range []string{"╭", "╰", "func println(args ...Type)", "The println built-in function docs."} {
		if !strings.Contains(got, want) {
			t.Errorf("janela de hover deveria conter %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "```") || strings.Contains(got, "https://") {
		t.Errorf("markdown cru vazou para a janela:\n%s", got)
	}

	typeKeys(t, ed, "<Esc>")
	if ed.hover != nil {
		t.Error("Esc deveria fechar o hover")
	}
	if ed.Cursor() != pos(3, 1) {
		t.Errorf("Esc que fecha o hover nao deveria mexer em nada; cursor %+v", ed.Cursor())
	}
}

func TestHoverAtrasadoEIgnorado(t *testing.T) {
	ed, fc := editorWithLSP(t, "a b\n")
	typeKeys(t, ed, "K")
	old := fc.lastID
	typeKeys(t, ed, "wK")
	ed.HandleAsync(lsp.HoverEvent{ID: old, Text: "resposta velha"})
	if ed.hover != nil {
		t.Error("resposta de um pedido antigo deveria ser ignorada")
	}
	ed.HandleAsync(lsp.HoverEvent{ID: fc.lastID, Text: ""})
	if !strings.Contains(ed.Frame().Status.Text(), "nada a mostrar") {
		t.Errorf("hover vazio deveria avisar: %q", ed.Frame().Status.Text())
	}
}

func TestHoverSemServidor(t *testing.T) {
	ed := editorFrom(t, "abc\n")
	typeKeys(t, ed, "K")
	if !strings.Contains(ed.Frame().Status.Text(), "sem servidor") {
		t.Errorf("status = %q", ed.Frame().Status.Text())
	}
}

func items(labels ...string) []lsp.CompletionItem {
	out := make([]lsp.CompletionItem, len(labels))
	for i, l := range labels {
		out[i] = lsp.CompletionItem{Label: l, Kind: 3, Detail: "func()"}
	}
	return out
}

func TestCompletion(t *testing.T) {
	ed, fc := editorWithLSP(t, "package main\n\nfunc main() {\n\t\n}\n")
	typeKeys(t, ed, "jjjAfmt.")
	if len(fc.comps) != 2 || fc.comps[0] != pos(3, 2) || fc.comps[1] != pos(3, 5) {
		t.Fatalf("completion pedido em %v; esperava so no inicio da palavra e depois do ponto", fc.comps)
	}
	ed.HandleAsync(lsp.CompletionEvent{ID: fc.lastID, Items: items("Println", "Printf", "Errorf", "Sprint")})
	if ed.comp == nil || len(ed.comp.items) != 4 {
		t.Fatalf("menu deveria abrir com 4 itens: %+v", ed.comp)
	}

	typeKeys(t, ed, "Pr")
	if len(fc.comps) != 2 {
		t.Errorf("com o menu aberto, digitar deveria so filtrar, nao pedir de novo (%d pedidos)", len(fc.comps))
	}
	if len(ed.comp.items) != 2 {
		t.Fatalf("filtrado por Pr deveria ter 2 itens, tem %d", len(ed.comp.items))
	}

	frame := ed.Frame()
	menu := ""
	for _, l := range frame.Lines {
		if strings.Contains(l.Text(), "Printf") {
			menu = l.Text()
		}
	}
	if !strings.Contains(menu, "func") {
		t.Errorf("menu deveria mostrar o tipo do item: %q", menu)
	}

	pressCtrl(ed, 'n')
	typeKeys(t, ed, "<CR>")
	if got := bufferText(ed); !strings.Contains(got, "\tfmt.Println\n") {
		t.Errorf("itens ficam em ordem alfabetica (Printf, Println); o segundo e Println: %q", got)
	}
	if ed.comp != nil || ed.Mode() != ModeInsert {
		t.Errorf("depois de aceitar, o menu fecha e continua no insert")
	}

	typeKeys(t, ed, "<Esc>u")
	if got := bufferText(ed); strings.Contains(got, "fmt") {
		t.Errorf("um u deveria desfazer a sessao de insert inteira, inclusive o completion: %q", got)
	}
}

func TestCompletionFechaEPedidosVelhos(t *testing.T) {
	ed, fc := editorWithLSP(t, "\n")
	typeKeys(t, ed, "ifo")
	first := fc.lastID
	ed.HandleAsync(lsp.CompletionEvent{ID: first, Items: items("foo", "format")})
	if ed.comp == nil {
		t.Fatal("menu deveria abrir")
	}
	typeKeys(t, ed, " ")
	if ed.comp != nil {
		t.Error("espaco deveria fechar o menu")
	}

	typeKeys(t, ed, "x")
	ed.HandleAsync(lsp.CompletionEvent{ID: first, Items: items("xyz")})
	if ed.comp != nil {
		t.Error("resposta de pedido antigo nao deveria abrir menu")
	}
	ed.HandleAsync(lsp.CompletionEvent{ID: fc.lastID, Items: items("abc")})
	if ed.comp != nil {
		t.Error("itens que nao batem com o prefixo nao deveriam abrir menu")
	}

	ed.HandleAsync(lsp.CompletionEvent{ID: fc.lastID, Items: items("xis")})
	if ed.comp == nil {
		t.Fatal("menu deveria abrir")
	}
	typeKeys(t, ed, "<Esc>")
	if ed.comp != nil || ed.Mode() != ModeNormal {
		t.Error("Esc fecha o menu e sai do insert")
	}
	ed.HandleAsync(lsp.CompletionEvent{ID: fc.lastID, Items: items("xis")})
	if ed.comp != nil {
		t.Error("no modo normal, resposta de completion nao abre menu")
	}
}

func TestCompletionManualComCtrlEspaco(t *testing.T) {
	ed, fc := editorWithLSP(t, "\n")
	typeKeys(t, ed, "i")
	pressCtrl(ed, ' ')
	if len(fc.comps) != 1 {
		t.Errorf("Ctrl+Espaco deveria pedir completion")
	}
}

func TestStatusDoServidor(t *testing.T) {
	ed := editorFrom(t, "a\n")
	fc := &fakeClient{events: make(chan lsp.Event, 1)}
	ed.SetLanguageClient(fc)
	if !strings.Contains(ed.Frame().Status.Text(), "gopls…") {
		t.Errorf("antes de ficar pronto: %q", ed.Frame().Status.Text())
	}
	ed.HandleAsync(lsp.StatusEvent{Err: errors.New("gopls não encontrado")})
	st := ed.Frame().Status
	if !strings.Contains(st.Text(), "gopls!") || !strings.Contains(st.Text(), "não encontrado") {
		t.Errorf("erro do servidor: %q", st.Text())
	}
	for _, s := range st {
		if strings.Contains(s.Text, "gopls!") && s.Style != styleStatusError {
			t.Errorf("indicador de erro deveria ser vermelho")
		}
	}
	if ed.Async() == nil {
		t.Error("Async deveria devolver o canal do cliente")
	}
	if editorFrom(t, "").Async() != nil {
		t.Error("sem cliente, Async deveria ser nil")
	}
}

func TestPopupNaoPassaDaTela(t *testing.T) {
	for _, size := range [][2]int{{20, 6}, {40, 10}, {80, 24}} {
		ed, fc := editorWithLSP(t, strings.Repeat("palavra\n", 30))
		ed.HandleEvent(input.NewResizeEvent(size[0], size[1]))
		typeKeys(t, ed, "GK")
		ed.HandleAsync(lsp.HoverEvent{ID: fc.lastID, Text: strings.Repeat("texto longo de documentacao ", 40)})
		for i, l := range ed.Frame().Lines {
			if l.Width() > size[0] {
				t.Errorf("tela %v: linha %d com largura %d", size, i, l.Width())
			}
		}
	}
	_ = render.Style{}
}

func TestHoverNaoCobreOCursor(t *testing.T) {
	for _, line := range []int{0, 4, 9} {
		ed, fc := editorWithLSP(t, strings.Repeat("palavra\n", 10))
		ed.HandleEvent(input.NewResizeEvent(60, 11))
		for i := 0; i < line; i++ {
			typeKeys(t, ed, "j")
		}
		typeKeys(t, ed, "K")
		ed.HandleAsync(lsp.HoverEvent{ID: fc.lastID, Text: strings.Repeat("linha\n\n", 30)})
		f := ed.Frame()
		if got := f.Lines[f.Cursor.Row].Text(); !strings.Contains(got, "palavra") || strings.Contains(got, "│") {
			t.Errorf("cursor na linha %d: a janela cobriu o cursor: %q", line, got)
		}
		boxRows := 0
		for _, l := range f.Lines {
			if strings.Contains(l.Text(), "│") || strings.Contains(l.Text(), "╭") || strings.Contains(l.Text(), "╰") {
				boxRows++
			}
		}
		if boxRows < 3 {
			t.Errorf("cursor na linha %d: a janela deveria aparecer cortada no lado com mais espaco (%d linhas)", line, boxRows)
		}
	}
}

func TestHoverTitulosMarkdown(t *testing.T) {
	h := newHoverPopup("### Printing\n\ntexto")
	if h == nil || !h.lines[0].heading || h.lines[0].text != "Printing" {
		t.Errorf("titulo markdown = %+v", h)
	}
}

func TestJanelaCortadaMostraReticencias(t *testing.T) {
	box := newHoverPopup(strings.Repeat("linha\n\n", 10)).render(20)
	small := shrinkBox(box, 5)
	if len(small) != 5 {
		t.Fatalf("shrinkBox(5) devolveu %d linhas", len(small))
	}
	if !strings.Contains(small[3].Text(), "…") || !strings.HasPrefix(small[4].Text(), "╰") {
		t.Errorf("janela cortada deveria terminar com … e a borda: %q / %q", small[3].Text(), small[4].Text())
	}
	if small[3].Width() != box[0].Width() {
		t.Errorf("linha de reticencias com largura %d, esperava %d", small[3].Width(), box[0].Width())
	}
}
