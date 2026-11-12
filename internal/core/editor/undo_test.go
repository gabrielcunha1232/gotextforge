package editor

import (
	"strings"
	"testing"
)

func TestUndoNoEditor(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		keys       string
		want       string
		wantCursor [2]int
		wantMsg    string
	}{
		{name: "u desfaz um x", src: "abc", keys: "lxu", want: "abc", wantCursor: [2]int{0, 1}},
		{name: "u desfaz a sessao de insert inteira", src: "fim", keys: "ihello<CR><Esc>u", want: "fim", wantCursor: [2]int{0, 0}},
		{name: "duas sessoes de insert sao dois passos", src: "", keys: "iab<Esc>acd<Esc>u", want: "ab", wantCursor: [2]int{0, 1}},
		{name: "u duas vezes desfaz as duas sessoes", src: "", keys: "iab<Esc>acd<Esc>uu", want: "", wantCursor: [2]int{0, 0}},
		{name: "o mais texto e um passo so", src: "abc\nxyz", keys: "onova<Esc>u", want: "abc\nxyz", wantCursor: [2]int{0, 0}},
		{name: "backspace que junta linhas e desfeito", src: "abc\nxyz", keys: "ji<BS><Esc>u", want: "abc\nxyz", wantCursor: [2]int{1, 0}},
		{name: "Ctrl+r refaz", src: "abc", keys: "x" + "u" + "\x00r", want: "bc", wantCursor: [2]int{0, 0}},
		{name: "movimentos nao viram passos de undo", src: "abc\nxyz", keys: "xjjkkllu", want: "abc\nxyz", wantCursor: [2]int{0, 0}},
		{name: "nada para desfazer avisa", src: "abc", keys: "u", want: "abc", wantMsg: "mais antiga"},
		{name: "nada para refazer avisa", src: "abc", keys: "\x00r", want: "abc", wantMsg: "mais recente"},
		{name: "cursor volta para onde estava antes do comando", src: "abc", keys: "A12345<Esc>$u", want: "abc", wantCursor: [2]int{0, 0}},
		{name: "cursor de antes do x volta junto", src: "abc", keys: "$xu", want: "abc", wantCursor: [2]int{0, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, tt.src)
			typeKeysWithCtrl(t, ed, tt.keys)

			if got := bufferText(ed); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if tt.wantMsg == "" {
				if got := ed.Cursor(); got != pos(tt.wantCursor[0], tt.wantCursor[1]) {
					t.Errorf("Cursor() = %+v, esperava %v", got, tt.wantCursor)
				}
			}
			if !strings.Contains(ed.Frame().Status, tt.wantMsg) {
				t.Errorf("status = %q, esperava conter %q", ed.Frame().Status, tt.wantMsg)
			}
			if ed.Mode() != ModeNormal {
				t.Errorf("modo = %v, esperava normal", ed.Mode())
			}
		})
	}
}

func typeKeysWithCtrl(t *testing.T, ed *Editor, keys string) {
	t.Helper()
	parts := strings.Split(keys, "\x00")
	typeKeys(t, ed, parts[0])
	for _, p := range parts[1:] {
		r := []rune(p)[0]
		pressCtrl(ed, r)
		typeKeys(t, ed, string([]rune(p)[1:]))
	}
}

func TestUndoTiraOMarcadorDeModificado(t *testing.T) {
	ed, _ := editorWithSaver(t, "abc\n")
	typeKeys(t, ed, "x")
	if !strings.Contains(ed.Frame().Status, "[+]") {
		t.Fatalf("depois de x deveria ter [+]: %q", ed.Frame().Status)
	}
	typeKeys(t, ed, "u")
	if strings.Contains(ed.Frame().Status, "[+]") {
		t.Errorf("depois de desfazer tudo nao deveria ter [+]: %q", ed.Frame().Status)
	}
	typeKeys(t, ed, ":q<CR>")
	if !ed.ShouldQuit() {
		t.Errorf(":q deveria sair, ja que nao ha alteracoes; status %q", ed.Frame().Status)
	}
}

func TestUndoDepoisDeSalvar(t *testing.T) {
	ed, fs := editorWithSaver(t, "abc\n")
	typeKeys(t, ed, "x:w<CR>u")
	if fs.calls != 1 {
		t.Fatalf("esperava 1 save, veio %d", fs.calls)
	}
	if !ed.buf.Modified() {
		t.Error("desfazer depois de salvar deixa o buffer diferente do disco")
	}
	if got := bufferText(ed); got != "abc" {
		t.Errorf("texto = %q, esperava %q", got, "abc")
	}
}
