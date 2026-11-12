package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
)

func styleOf(t *testing.T, line render.StyledLine, word string) render.Style {
	t.Helper()
	for _, s := range line {
		if strings.Contains(s.Text, word) {
			return s.Style
		}
	}
	t.Fatalf("%q nao aparece na linha %+v", word, line)
	return render.Style{}
}

func TestHighlightNoFrame(t *testing.T) {
	ed := editorFrom(t, "package main\n\nfunc main() {\n\tprintln(\"oi\", 42) // fim\n}\n")
	f := ed.Frame()

	checks := []struct {
		line int
		word string
		kind syntax.Kind
	}{
		{0, "package", syntax.KindKeyword},
		{2, "func", syntax.KindKeyword},
		{2, "main", syntax.KindFunc},
		{3, "println", syntax.KindPredeclared},
		{3, "\"oi\"", syntax.KindString},
		{3, "42", syntax.KindNumber},
		{3, "// fim", syntax.KindComment},
	}
	for _, c := range checks {
		if got := withBg(styleOf(t, f.Lines[c.line], c.word), render.Color{}); got != theme[c.kind] {
			t.Errorf("linha %d %q: estilo %+v, esperava %+v", c.line, c.word, got, theme[c.kind])
		}
	}
}

func TestHighlightSoEmArquivoGo(t *testing.T) {
	buf, err := buffer.Load(strings.NewReader("package main // nada\n"))
	if err != nil {
		t.Fatal(err)
	}
	ed := New(buf, "README.md")
	ed.HandleEvent(input.NewResizeEvent(80, 24))
	for _, s := range ed.Frame().Lines[0][1:] {
		if s.Style.Fg != (render.Color{}) {
			t.Errorf("arquivo que nao e .go nao deveria ter cores: %+v", s)
		}
	}
}

func TestHighlightAcompanhaEdicao(t *testing.T) {
	ed := editorFrom(t, "x := 1\ny := 2\nz := 3\n")
	comment := theme[syntax.KindComment]

	if got := styleOf(t, ed.Frame().Lines[2], "z"); got == comment {
		t.Fatal("antes da edicao a linha 3 nao e comentario")
	}

	typeKeys(t, ed, "A /* abre<Esc>")
	f := ed.Frame()
	for i, word := range []string{"y := 2", "z := 3"} {
		if got := styleOf(t, f.Lines[i+1], word); got != comment {
			t.Errorf("depois de abrir /* a linha %d deveria ser comentario, veio %+v", i+2, got)
		}
	}

	typeKeys(t, ed, "u")
	if got := styleOf(t, ed.Frame().Lines[2], "z"); got == comment {
		t.Error("depois do undo a linha 3 deveria voltar a ser codigo")
	}
}

func TestHighlightComTelaRolada(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("var q = `inicio\n")
	for i := 0; i < 200; i++ {
		sb.WriteString("conteudo da raw string\n")
	}
	sb.WriteString("fim`\nfunc depois() {}\n")
	ed := editorFrom(t, sb.String())

	typeKeys(t, ed, "G")
	f := ed.Frame()
	last := len(f.Lines) - 1
	for last > 0 && strings.HasPrefix(frameLineText(f.Lines[last]), "~") {
		last--
	}
	if got := withBg(styleOf(t, f.Lines[last-2], "conteudo"), render.Color{}); got != theme[syntax.KindString] {
		t.Errorf("linha no meio da raw string, vista depois de G, deveria ser string; veio %+v", got)
	}
	if got := withBg(styleOf(t, f.Lines[last], "func"), render.Color{}); got != theme[syntax.KindKeyword] {
		t.Errorf("depois da raw string fechar, func deveria ser keyword; veio %+v", got)
	}
}
