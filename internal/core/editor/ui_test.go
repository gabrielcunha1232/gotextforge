package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func gutterText(line render.StyledLine) string {
	if len(line) == 0 {
		return ""
	}
	return line[0].Text
}

func TestGutterNumerosRelativos(t *testing.T) {
	ed := editorFrom(t, strings.Repeat("x\n", 20))
	typeKeys(t, ed, "jjjjj")
	f := ed.Frame()

	tests := []struct {
		row  int
		want string
	}{
		{row: 0, want: "   5 "},
		{row: 4, want: "   1 "},
		{row: 5, want: " 6   "},
		{row: 6, want: "   1 "},
		{row: 10, want: "   5 "},
	}
	for _, tt := range tests {
		if got := gutterText(f.Lines[tt.row]); got != tt.want {
			t.Errorf("gutter da linha %d = %q, esperava %q", tt.row, got, tt.want)
		}
	}
	if f.Lines[5][0].Style != styleCursorLineNr {
		t.Errorf("numero da linha atual deveria usar styleCursorLineNr")
	}
}

func TestGutterCresceComOArquivo(t *testing.T) {
	ed := editorFrom(t, strings.Repeat("x\n", 12345))
	if got := ed.gutterWidth(); got != 7 {
		t.Errorf("gutterWidth com 12345 linhas = %d, esperava 7 (5 digitos + 2)", got)
	}
}

func TestLinhaDoCursorOcupaALarguraToda(t *testing.T) {
	ed := editorFrom(t, "curta\noutra\n")
	f := ed.Frame()
	if w := f.Lines[0].Width(); w != 80 {
		t.Errorf("linha do cursor tem largura %d, esperava 80", w)
	}
	for _, s := range f.Lines[0][1:] {
		if s.Style.Bg != colBgHighlight {
			t.Errorf("span %q da linha do cursor sem o fundo de destaque", s.Text)
		}
	}
	if w := f.Lines[1].Width(); w == 80 {
		t.Error("outras linhas nao deveriam ser preenchidas ate a borda")
	}
}

func TestStatusLineLarguraEmVariasTelas(t *testing.T) {
	for _, width := range []int{1, 5, 10, 20, 40, 80, 200} {
		ed := editorFrom(t, "abc\n")
		ed.HandleEvent(input.NewResizeEvent(width, 10))
		typeKeys(t, ed, "x")
		got := ed.Frame().Status
		if w := got.Width(); w != width {
			t.Errorf("largura %d: status ocupa %d celulas (%q)", width, w, got.Text())
		}
	}
}

func TestStatusLineMostraErroEmVermelho(t *testing.T) {
	ed := editorFrom(t, "abc\n")
	typeKeys(t, ed, ":nada<CR>")
	f := ed.Frame()
	found := false
	for _, s := range f.Status {
		if strings.Contains(s.Text, "E492") {
			found = true
			if s.Style != styleStatusError {
				t.Errorf("mensagem de erro com estilo %+v, esperava styleStatusError", s.Style)
			}
		}
	}
	if !found {
		t.Fatalf("mensagem E492 nao aparece: %q", f.Status.Text())
	}
}

func TestStatusLineNeutralizaNomeEstranho(t *testing.T) {
	ed := editorFrom(t, "abc\n")
	ed.filename = "a\x1b[2Jb.go"
	if got := ed.Frame().Status.Text(); strings.ContainsRune(got, 0x1b) {
		t.Errorf("status contem ESC cru: %q", got)
	}
}

func TestIntroAparecePorqueNaoHaArquivo(t *testing.T) {
	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))

	text := func() string {
		var sb strings.Builder
		for _, l := range ed.Frame().Lines {
			sb.WriteString(l.Text())
			sb.WriteByte('\n')
		}
		return sb.String()
	}

	got := text()
	for _, want := range []string{"█████", "editor modal para Go", "começar a escrever"} {
		if !strings.Contains(got, want) {
			t.Errorf("tela inicial deveria conter %q", want)
		}
	}

	typeKeys(t, ed, "iola<Esc>")
	if strings.Contains(text(), "editor modal para Go") {
		t.Error("a tela inicial deveria sumir depois de digitar")
	}
}

func TestIntroNaoApareceComArquivo(t *testing.T) {
	ed := editorFrom(t, "")
	for _, l := range ed.Frame().Lines {
		if strings.Contains(l.Text(), "editor modal") {
			t.Fatal("arquivo aberto (mesmo vazio) nao deveria mostrar a tela inicial")
		}
	}
}

func TestIntroEmTelasPequenasNaoQuebra(t *testing.T) {
	for _, size := range [][2]int{{1, 2}, {10, 5}, {30, 12}, {60, 20}, {66, 40}, {67, 40}} {
		ed := New(nil, "")
		ed.HandleEvent(input.NewResizeEvent(size[0], size[1]))
		f := ed.Frame()
		for i, l := range f.Lines {
			if w := l.Width(); w > size[0] {
				t.Errorf("tela %dx%d: linha %d tem largura %d", size[0], size[1], i, w)
			}
		}
	}
}
