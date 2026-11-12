package view

import (
	"testing"
	"unicode/utf8"
)

func TestScreenCol(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		runeCol int
		want    int
	}{
		{name: "inicio da linha", line: "abc", runeCol: 0, want: 0},
		{name: "texto simples", line: "abc", runeCol: 2, want: 2},
		{name: "depois de uma tab", line: "\tx", runeCol: 1, want: 4},
		{name: "depois de duas tabs", line: "\t\tx := 1", runeCol: 2, want: 8},
		{name: "tab no meio alinha na parada", line: "ab\tc", runeCol: 3, want: 4},
		{name: "controle ocupa 2 celulas", line: "a\rb", runeCol: 2, want: 3},
		{name: "C1 ocupa 4 celulas", line: "a\u009bb", runeCol: 2, want: 5},
		{name: "acentos ocupam 1 celula", line: "ação", runeCol: 3, want: 3},
		{name: "coluna alem do fim para no fim", line: "ab", runeCol: 10, want: 2},
		{name: "linha vazia", line: "", runeCol: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScreenCol([]rune(tt.line), tt.runeCol, DefaultTabStop); got != tt.want {
				t.Errorf("ScreenCol(%q, %d) = %d, esperava %d", tt.line, tt.runeCol, got, tt.want)
			}
		})
	}
}

func TestScreenColConcordaComFormatLine(t *testing.T) {
	lines := []string{
		"",
		"func main() {",
		"\tfmt.Println(\"oi\")",
		"\t\tif x {\treturn }",
		"ab\tcd\tef",
		"a\rb\x1b[2Jc\x7f",
		"x\u009by\u0080z",
		"ação\té\tfim",
	}

	for _, line := range lines {
		runes := []rune(line)
		for c := 0; c <= len(runes); c++ {
			formatted := FormatLine(runes[:c], 1<<20, DefaultTabStop)
			want := utf8.RuneCountInString(formatted)
			if got := ScreenCol(runes, c, DefaultTabStop); got != want {
				t.Errorf("linha %q, runeCol %d: ScreenCol = %d, FormatLine desenha %d celulas", line, c, got, want)
			}
		}
	}
}

func TestEnsureVisible(t *testing.T) {
	tests := []struct {
		name       string
		top        int
		height     int
		totalLines int
		line       int
		wantTop    int
	}{
		{name: "ja visivel nao mexe", top: 10, height: 10, totalLines: 100, line: 15, wantTop: 10},
		{name: "primeira linha visivel nao mexe", top: 10, height: 10, totalLines: 100, line: 10, wantTop: 10},
		{name: "ultima linha visivel nao mexe", top: 10, height: 10, totalLines: 100, line: 19, wantTop: 10},
		{name: "uma abaixo da tela rola 1", top: 10, height: 10, totalLines: 100, line: 20, wantTop: 11},
		{name: "uma acima da tela rola 1", top: 10, height: 10, totalLines: 100, line: 9, wantTop: 9},
		{name: "salto para baixo deixa a linha no rodape", top: 0, height: 10, totalLines: 100, line: 50, wantTop: 41},
		{name: "salto para cima deixa a linha no topo", top: 80, height: 10, totalLines: 100, line: 3, wantTop: 3},
		{name: "ultima linha do arquivo", top: 0, height: 10, totalLines: 100, line: 99, wantTop: 90},
		{name: "arquivo menor que a tela", top: 0, height: 30, totalLines: 5, line: 4, wantTop: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := Viewport{Top: tt.top, Height: tt.height}
			vp.EnsureVisible(tt.line, tt.totalLines)
			if vp.Top != tt.wantTop {
				t.Errorf("Top = %d, esperava %d", vp.Top, tt.wantTop)
			}
		})
	}
}
