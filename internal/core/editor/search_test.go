package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func TestBusca(t *testing.T) {
	const src = "func main() {\n\terr := f()\n\tif err != nil {\n\t\treturn err\n\t}\n}\n"
	tests := []struct {
		name    string
		keys    string
		want    [2]int
		wantMsg string
	}{
		{name: "/ vai para a primeira ocorrencia", keys: "/err<CR>", want: [2]int{1, 1}, wantMsg: "/err"},
		{name: "n vai para a proxima", keys: "/err<CR>n", want: [2]int{2, 4}},
		{name: "3n", keys: "/err<CR>3n", want: [2]int{1, 1}, wantMsg: "FIM"},
		{name: "N volta", keys: "/err<CR>nN", want: [2]int{1, 1}},
		{name: "n volta ao inicio", keys: "/err<CR>nnn", want: [2]int{1, 1}, wantMsg: "continuou do INÍCIO"},
		{name: "? busca para tras", keys: "G?err<CR>", want: [2]int{3, 9}},
		{name: "n depois de ? continua para tras", keys: "G?err<CR>n", want: [2]int{2, 4}},
		{name: "/ vazio repete a ultima busca", keys: "/err<CR>/<CR>", want: [2]int{2, 4}},
		{name: "maiuscula nao acha minuscula", keys: "/Err<CR>", want: [2]int{0, 0}, wantMsg: "E486"},
		{name: "nao encontrado", keys: "/zzz<CR>", want: [2]int{0, 0}, wantMsg: "E486: padrão não encontrado: zzz"},
		{name: "n sem busca anterior", keys: "n", want: [2]int{0, 0}, wantMsg: "E35"},
		{name: "Esc cancela a busca", keys: "/err<Esc>", want: [2]int{0, 0}},
		{name: "Backspace corrige o padrao", keys: "/errx<BS><CR>", want: [2]int{1, 1}},
		{name: "* busca a palavra sob o cursor", keys: "jw*", want: [2]int{2, 4}},
		{name: "* depois do fim das palavras da linha avisa", keys: "j$h*", want: [2]int{1, 9}, wantMsg: "E348"},
		{name: "# para tras", keys: "jjj$#", want: [2]int{2, 4}},
		{name: "* usa palavra inteira", keys: "w*", want: [2]int{0, 5}, wantMsg: "continuou"},
		{name: "* em linha sem palavra", keys: "G*", want: [2]int{5, 0}, wantMsg: "E348"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, src)
			typeKeys(t, ed, tt.keys)
			if got := ed.Cursor(); got != pos(tt.want[0], tt.want[1]) {
				t.Errorf("cursor = %+v, esperava %v", got, tt.want)
			}
			if ed.Mode() != ModeNormal {
				t.Errorf("modo = %v", ed.Mode())
			}
			if st := ed.Frame().Status.Text(); !strings.Contains(st, tt.wantMsg) {
				t.Errorf("status = %q, esperava conter %q", st, tt.wantMsg)
			}
		})
	}
}

func searchStyles(f render.Frame, row int) (normal, current int) {
	for _, s := range f.Lines[row] {
		switch {
		case s.Style.Bg == styleSearch.Bg:
			normal += len([]rune(s.Text))
		case s.Style.Bg == styleSearchCurrent.Bg:
			current += len([]rune(s.Text))
		}
	}
	return normal, current
}

func TestBuscaDestaca(t *testing.T) {
	ed := editorFrom(t, "a err b err\nerr\n")
	typeKeys(t, ed, "/er")
	f := ed.Frame()
	if n, _ := searchStyles(f, 0); n != 4 {
		t.Errorf("enquanto digita, as 2 ocorrencias de 'er' na linha 1 deveriam estar destacadas (%d celulas)", n)
	}
	if f.Status.Text() != "/er" {
		t.Errorf("status durante a busca = %q", f.Status.Text())
	}
	if f.Cursor.Row != len(f.Lines) || f.Cursor.Col != 3 {
		t.Errorf("cursor durante a busca deveria estar na linha de status, depois de /er: %+v", f.Cursor)
	}

	typeKeys(t, ed, "r<CR>")
	f = ed.Frame()
	n, cur := searchStyles(f, 0)
	if cur != 3 || n != 3 {
		t.Errorf("depois de buscar: ocorrencia atual %d celulas (esperava 3), outras %d (esperava 3)", cur, n)
	}
	if n, _ := searchStyles(f, 1); n != 3 {
		t.Errorf("a ocorrencia da linha 2 deveria estar destacada")
	}

	typeKeys(t, ed, ":noh<CR>")
	f = ed.Frame()
	if n, cur := searchStyles(f, 0); n+cur != 0 {
		t.Errorf(":noh deveria apagar o destaque")
	}
	typeKeys(t, ed, "n")
	if n, cur := searchStyles(ed.Frame(), 0); n+cur == 0 {
		t.Errorf("n deveria reativar o destaque")
	}
}
