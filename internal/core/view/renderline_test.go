package view

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

var (
	sA = render.Style{Fg: render.ColorBlue}
	sB = render.Style{Fg: render.ColorGreen}
)

func styleByMask(mask string) func(int) render.Style {
	return func(i int) render.Style {
		if i < len(mask) {
			switch mask[i] {
			case 'a':
				return sA
			case 'b':
				return sB
			}
		}
		return render.Style{}
	}
}

func describeLine(line render.StyledLine) string {
	names := map[render.Style]string{{}: "_", sA: "a", sB: "b"}
	parts := make([]string, len(line))
	for i, s := range line {
		parts[i] = fmt.Sprintf("%s[%s]", names[s.Style], s.Text)
	}
	return strings.Join(parts, " ")
}

func TestRenderLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		mask  string
		width int
		want  string
	}{
		{name: "linha vazia nao gera spans", line: "", mask: "", width: 80, want: ""},
		{name: "tudo sem estilo vira 1 span", line: "abc", mask: "___", width: 80, want: "_[abc]"},
		{name: "estilos vizinhos iguais se juntam", line: "func main", mask: "aaaa_bbbb", width: 80, want: "a[func] _[ ] b[main]"},
		{name: "tab herda o estilo da rune", line: "\tx", mask: "ab", width: 80, want: "a[    ] b[x]"},
		{name: "controle herda o estilo", line: "a\rb", mask: "_a_", width: 80, want: "_[a] a[^M] _[b]"},
		{name: "corte na largura", line: "abcdef", mask: "aaabbb", width: 4, want: "a[abc] b[d]"},
		{name: "corte no meio da tab", line: "a\tb", mask: "_a_", width: 3, want: "_[a] a[  ]"},
		{name: "acentos", line: "ação", mask: "aabb", width: 80, want: "a[aç] b[ão]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderLine([]rune(tt.line), styleByMask(tt.mask), tt.width, DefaultTabStop)
			if d := describeLine(got); d != tt.want {
				t.Errorf("RenderLine = %s\nesperava     %s", d, tt.want)
			}
		})
	}
}

func TestRenderLineConcordaComFormatLine(t *testing.T) {
	lines := []string{"", "\t\tif x {\treturn }", "a\rb\x1b[2Jc\x7f", "x\u009by", "ação\té"}
	for _, line := range lines {
		for _, width := range []int{1, 3, 7, 80} {
			got := RenderLine([]rune(line), styleByMask("abab_abab_ab"), width, DefaultTabStop)
			var sb strings.Builder
			for _, s := range got {
				sb.WriteString(s.Text)
			}
			if want := FormatLine([]rune(line), width, DefaultTabStop); sb.String() != want {
				t.Errorf("linha %q largura %d: RenderLine desenha %q, FormatLine desenha %q", line, width, sb.String(), want)
			}
		}
	}
}
