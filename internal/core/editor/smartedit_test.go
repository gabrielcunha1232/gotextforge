package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TestAutoFechamentoEIndentacao(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		keys    string
		want    string
		wantPos [2]int
	}{
		{name: "( fecha sozinho", src: "", keys: "i(", want: "()", wantPos: [2]int{0, 1}},
		{name: "[ e { fecham", src: "", keys: "ix := []int{", want: "x := []int{}", wantPos: [2]int{0, 11}},
		{name: "digitar o fechamento pula por cima", src: "", keys: "if(x)", want: "f(x)", wantPos: [2]int{0, 4}},
		{name: "aninhado", src: "", keys: "if(g(1))", want: "f(g(1))", wantPos: [2]int{0, 7}},
		{name: "nao fecha antes de uma palavra", src: "abc", keys: "i(", want: "(abc", wantPos: [2]int{0, 1}},
		{name: "fecha antes de espaco", src: " abc", keys: "i(", want: "() abc", wantPos: [2]int{0, 1}},
		{name: "aspas fecham", src: "", keys: "is := \"", want: "s := \"\"", wantPos: [2]int{0, 6}},
		{name: "aspas: digitar a de fechar pula", src: "", keys: "i\"oi\"", want: "\"oi\"", wantPos: [2]int{0, 4}},
		{name: "apostrofo depois de letra nao fecha", src: "", keys: "idon't", want: "don't", wantPos: [2]int{0, 5}},
		{name: "aspas depois de barra nao fecham", src: "", keys: "i\\\"", want: "\\\"", wantPos: [2]int{0, 2}},
		{name: "rune literal", src: "", keys: "ic := 'x", want: "c := 'x'", wantPos: [2]int{0, 7}},
		{name: "backspace entre par vazio apaga os dois", src: "", keys: "if(<BS>", want: "f", wantPos: [2]int{0, 1}},
		{name: "backspace entre aspas vazias", src: "", keys: "i\"<BS>", want: "", wantPos: [2]int{0, 0}},
		{name: "backspace com conteudo apaga so um", src: "", keys: "i(x<BS><BS>", want: "", wantPos: [2]int{0, 0}},

		{name: "Enter mantem a indentacao", src: "\t\tfoo", keys: "A<CR>bar", want: "\t\tfoo\n\t\tbar", wantPos: [2]int{1, 5}},
		{name: "Enter depois de { indenta mais", src: "\tif x {", keys: "A<CR>y", want: "\tif x {\n\t\ty", wantPos: [2]int{1, 3}},
		{name: "Enter entre {} abre o bloco", src: "", keys: "ifunc f() {<CR>return", want: "func f() {\n\treturn\n}", wantPos: [2]int{1, 7}},
		{name: "Enter entre () abre", src: "", keys: "if(<CR>a,", want: "f(\n\ta,\n)", wantPos: [2]int{1, 3}},
		{name: "Enter em linha so com indentacao limpa os espacos", src: "\tif x {", keys: "A<CR><CR>z", want: "\tif x {\n\n\t\tz", wantPos: [2]int{2, 3}},
		{name: "} em linha vazia alinha com a {", src: "\tif x {\n\t\ty", keys: "jA<CR>}", want: "\tif x {\n\t\ty\n\t}", wantPos: [2]int{2, 2}},
		{name: "o depois de { indenta", src: "func f() {\n}", keys: "ox", want: "func f() {\n\tx\n}", wantPos: [2]int{1, 2}},
		{name: "o mantem a indentacao", src: "\t\tfoo()", keys: "ox", want: "\t\tfoo()\n\t\tx", wantPos: [2]int{1, 3}},
		{name: "O acima de } indenta para dentro", src: "func f() {\n}", keys: "jOx", want: "func f() {\n\tx\n}", wantPos: [2]int{1, 2}},
		{name: "um u desfaz o bloco inteiro", src: "", keys: "ifunc f() {<CR>return<Esc>u", want: "", wantPos: [2]int{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, tt.src)
			typeKeys(t, ed, tt.keys)
			if got := bufferText(ed); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if !strings.HasSuffix(tt.keys, "u") && !strings.HasSuffix(tt.keys, "<Esc>") {
				if got := ed.Cursor(); got != pos(tt.wantPos[0], tt.wantPos[1]) {
					t.Errorf("cursor = %+v, esperava %v", got, tt.wantPos)
				}
			}
		})
	}
}

func TestIndentacaoComEspacosForaDoGo(t *testing.T) {
	buf, err := buffer.Load(strings.NewReader("    if x:"))
	if err != nil {
		t.Fatal(err)
	}
	ed := New(buf, "script.py")
	ed.HandleEvent(input.NewResizeEvent(80, 24))
	typeKeys(t, ed, "A(<CR>y")
	if got := bufferText(ed); got != "    if x:(\n        y\n    )" {
		t.Errorf("fora do Go, a indentacao com espacos usa 4 espacos por nivel: %q", got)
	}
}
