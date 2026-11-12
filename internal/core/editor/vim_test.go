package editor

import (
	"strings"
	"testing"
)

func TestVim(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		keys     string
		want     string
		wantPos  [2]int
		wantMode Mode
	}{
		{name: "dd", src: "a\nb\nc", keys: "jdd", want: "a\nc", wantPos: [2]int{1, 0}},
		{name: "3dd", src: "a\nb\nc\nd\ne", keys: "j3dd", want: "a\ne", wantPos: [2]int{1, 0}},
		{name: "dd na ultima linha sobe o cursor", src: "a\nb\nc", keys: "Gdd", want: "a\nb", wantPos: [2]int{1, 0}},
		{name: "dd na unica linha deixa o buffer vazio", src: "abc", keys: "dd", want: "", wantPos: [2]int{0, 0}},
		{name: "dd vai para o primeiro nao-branco", src: "x\n\tfoo\nbar", keys: "dd", want: "\tfoo\nbar", wantPos: [2]int{0, 1}},
		{name: "dw", src: "foo bar baz", keys: "dw", want: "bar baz", wantPos: [2]int{0, 0}},
		{name: "d2w", src: "foo bar baz", keys: "d2w", want: "baz", wantPos: [2]int{0, 0}},
		{name: "2dw", src: "foo bar baz", keys: "2dw", want: "baz", wantPos: [2]int{0, 0}},
		{name: "dw na ultima palavra da linha nao junta linhas", src: "foo bar\nbaz", keys: "wdw", want: "foo \nbaz", wantPos: [2]int{0, 3}},
		{name: "de", src: "foo bar", keys: "de", want: " bar", wantPos: [2]int{0, 0}},
		{name: "db", src: "foo bar", keys: "$db", want: "foo r", wantPos: [2]int{0, 4}},
		{name: "d$", src: "foo bar", keys: "wd$", want: "foo ", wantPos: [2]int{0, 3}},
		{name: "D", src: "foo bar", keys: "wD", want: "foo ", wantPos: [2]int{0, 3}},
		{name: "d0", src: "foo bar", keys: "wd0", want: "bar", wantPos: [2]int{0, 0}},
		{name: "dfx inclui o x", src: "abcxdef", keys: "dfx", want: "def", wantPos: [2]int{0, 0}},
		{name: "dtx nao inclui o x", src: "abcxdef", keys: "dtx", want: "xdef", wantPos: [2]int{0, 0}},
		{name: "dj apaga duas linhas", src: "a\nb\nc", keys: "dj", want: "c", wantPos: [2]int{0, 0}},
		{name: "dk", src: "a\nb\nc", keys: "Gdk", want: "a", wantPos: [2]int{0, 0}},
		{name: "dG", src: "a\nb\nc", keys: "jdG", want: "a", wantPos: [2]int{0, 0}},
		{name: "dgg", src: "a\nb\nc", keys: "jdgg", want: "c", wantPos: [2]int{0, 0}},
		{name: "dj na ultima linha nao faz nada", src: "a\nb", keys: "jdj", want: "a\nb", wantPos: [2]int{1, 0}},
		{name: "x com count", src: "abcdef", keys: "l3x", want: "aef", wantPos: [2]int{0, 1}},
		{name: "X", src: "abc", keys: "$X", want: "ac", wantPos: [2]int{0, 1}},
		{name: "ciw", src: "foo bar baz", keys: "wciwXYZ<Esc>", want: "foo XYZ baz", wantPos: [2]int{0, 6}},
		{name: "ciw fica no insert", src: "foo bar", keys: "ciwX", want: "X bar", wantPos: [2]int{0, 1}, wantMode: ModeInsert},
		{name: "daw", src: "foo bar baz", keys: "wdaw", want: "foo baz", wantPos: [2]int{0, 4}},
		{name: "cw muda so ate o fim da palavra", src: "foo bar", keys: "cwX<Esc>", want: "X bar", wantPos: [2]int{0, 0}},
		{name: "cw no ultimo caractere da palavra", src: "foo bar", keys: "llcwX<Esc>", want: "foX bar", wantPos: [2]int{0, 2}},
		{name: "c2w", src: "a b c", keys: "c2wX<Esc>", want: "X c", wantPos: [2]int{0, 0}},
		{name: "ci\"", src: "s := \"velho\"", keys: "fvci\"novo<Esc>", want: "s := \"novo\"", wantPos: [2]int{0, 9}},
		{name: "di(", src: "f(a, b)", keys: "fadi(", want: "f()", wantPos: [2]int{0, 2}},
		{name: "da(", src: "x f(a) y", keys: "fada(", want: "x f y", wantPos: [2]int{0, 3}},
		{name: "di{ multilinha apaga as linhas de dentro", src: "func f() {\n\ta()\n\tb()\n}", keys: "jdi{", want: "func f() {\n}", wantPos: [2]int{1, 0}},
		{name: "ci{ multilinha deixa uma linha vazia", src: "func f() {\n\ta()\n\tb()\n}", keys: "jci{X<Esc>", want: "func f() {\nX\n}", wantPos: [2]int{1, 0}},
		{name: "dap", src: "a\nb\n\nc\nd", keys: "dap", want: "c\nd", wantPos: [2]int{0, 0}},
		{name: "dap no ultimo paragrafo", src: "a\n\nc\nd", keys: "Gdap", want: "a", wantPos: [2]int{0, 0}},
		{name: "dip", src: "a\nb\n\nc", keys: "dip", want: "\nc", wantPos: [2]int{0, 0}},
		{name: "cc", src: "a\nfoo\nb", keys: "jccX<Esc>", want: "a\nX\nb", wantPos: [2]int{1, 0}},
		{name: "C", src: "foo bar", keys: "wCX<Esc>", want: "foo X", wantPos: [2]int{0, 4}},
		{name: "S", src: "foo", keys: "SX<Esc>", want: "X", wantPos: [2]int{0, 0}},
		{name: "s", src: "foo", keys: "sX<Esc>", want: "Xoo", wantPos: [2]int{0, 0}},
		{name: "yy p", src: "a\nb", keys: "yyp", want: "a\na\nb", wantPos: [2]int{1, 0}},
		{name: "yy P", src: "a\nb", keys: "jyyP", want: "a\nb\nb", wantPos: [2]int{1, 0}},
		{name: "3p", src: "a", keys: "yy3p", want: "a\na\na\na", wantPos: [2]int{1, 0}},
		{name: "yw p", src: "foo bar", keys: "ywP", want: "foo foo bar", wantPos: [2]int{0, 3}},
		{name: "dd p troca linhas", src: "a\nb\nc", keys: "ddp", want: "b\na\nc", wantPos: [2]int{1, 0}},
		{name: "x p troca letras", src: "ab", keys: "xp", want: "ba", wantPos: [2]int{0, 1}},
		{name: "dw P desfaz", src: "foo bar", keys: "dwP", want: "foo bar", wantPos: [2]int{0, 3}},
		{name: "p sem nada copiado avisa", src: "abc", keys: "p", want: "abc", wantPos: [2]int{0, 0}},
		{name: "yj copia duas linhas", src: "a\nb\nc", keys: "yjGp", want: "a\nb\nc\na\nb", wantPos: [2]int{3, 0}},
		{name: "J", src: "a\n   b", keys: "J", want: "a b", wantPos: [2]int{0, 1}},
		{name: "3J", src: "a\nb\nc\nd", keys: "3J", want: "a b c\nd", wantPos: [2]int{0, 3}},
		{name: "J com linha vazia", src: "a\n\nb", keys: "J", want: "a\nb", wantPos: [2]int{0, 0}},
		{name: "rx", src: "abc", keys: "lrx", want: "axc", wantPos: [2]int{0, 1}},
		{name: "3rx", src: "abcd", keys: "3rx", want: "xxxd", wantPos: [2]int{0, 2}},
		{name: "r alem do fim nao faz nada", src: "ab", keys: "5rx", want: "ab", wantPos: [2]int{0, 0}},
		{name: "5j", src: "1\n2\n3\n4\n5\n6\n7", keys: "5j", want: "1\n2\n3\n4\n5\n6\n7", wantPos: [2]int{5, 0}},
		{name: "3G", src: "1\n2\n3\n4", keys: "3G", want: "1\n2\n3\n4", wantPos: [2]int{2, 0}},
		{name: "w e b", src: "foo bar baz", keys: "wwbe", want: "foo bar baz", wantPos: [2]int{0, 6}},
		{name: "f e ;-ausente", src: "a(b)c", keys: "f)", want: "a(b)c", wantPos: [2]int{0, 3}},
		{name: "u desfaz dd inteiro", src: "a\nb\nc", keys: "j2ddu", want: "a\nb\nc", wantPos: [2]int{1, 0}},
		{name: "u desfaz ciw mais o texto digitado", src: "foo bar", keys: "ciwXYZ<Esc>u", want: "foo bar", wantPos: [2]int{0, 0}},
		{name: "3u", src: "abcd", keys: "xxx3u", want: "abcd", wantPos: [2]int{0, 0}},
		{name: "Esc cancela operador pendente", src: "abc", keys: "d<Esc>x", want: "bc", wantPos: [2]int{0, 0}},
		{name: "I vai ao primeiro nao-branco", src: "\t\tfoo", keys: "$IX<Esc>", want: "\t\tXfoo", wantPos: [2]int{0, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, tt.src)
			typeKeys(t, ed, tt.keys)
			if got := bufferText(ed); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if got := ed.Cursor(); got != pos(tt.wantPos[0], tt.wantPos[1]) {
				t.Errorf("cursor = %+v, esperava %v", got, tt.wantPos)
			}
			if ed.Mode() != tt.wantMode {
				t.Errorf("modo = %v, esperava %v", ed.Mode(), tt.wantMode)
			}
		})
	}
}

func TestShowcmd(t *testing.T) {
	ed := editorFrom(t, "abc\n")
	typeKeys(t, ed, "2d")
	if got := ed.Frame().Status.Text(); !strings.Contains(got, " 2d ") {
		t.Errorf("status deveria mostrar o comando pendente 2d: %q", got)
	}
	typeKeys(t, ed, "<Esc>")
	if got := ed.Frame().Status.Text(); strings.Contains(got, "2d") {
		t.Errorf("depois do Esc o pendente deveria sumir: %q", got)
	}
}

func TestMensagensDeOperadores(t *testing.T) {
	ed := editorFrom(t, "a\nb\nc\nd\n")
	typeKeys(t, ed, "3yy")
	if got := ed.Frame().Status.Text(); !strings.Contains(got, "3 linhas copiadas") {
		t.Errorf("status = %q", got)
	}
	typeKeys(t, ed, "3dd")
	if got := ed.Frame().Status.Text(); !strings.Contains(got, "3 linhas a menos") {
		t.Errorf("status = %q", got)
	}
	ed2 := editorFrom(t, "abc\n")
	typeKeys(t, ed2, "p")
	if got := ed2.Frame().Status.Text(); !strings.Contains(got, "E353") {
		t.Errorf("p sem registrador deveria avisar E353: %q", got)
	}
}
