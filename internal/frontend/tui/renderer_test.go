package tui

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func TestWriteSpan(t *testing.T) {
	tests := []struct {
		name  string
		style render.Style
		want  string
	}{
		{name: "sem estilo nao gera escape", style: render.Style{}, want: "x"},
		{name: "ANSI basico", style: render.Style{Fg: render.ColorMagenta}, want: "\x1b[35mx\x1b[0m"},
		{name: "ANSI brilhante", style: render.Style{Fg: render.ColorBrightBlack}, want: "\x1b[90mx\x1b[0m"},
		{name: "paleta 256", style: render.Style{Fg: render.ANSI(208)}, want: "\x1b[38;5;208mx\x1b[0m"},
		{name: "RGB", style: render.Style{Fg: render.RGB(255, 85, 0)}, want: "\x1b[38;2;255;85;0mx\x1b[0m"},
		{name: "fundo", style: render.Style{Bg: render.ColorBlue}, want: "\x1b[44mx\x1b[0m"},
		{name: "fundo brilhante", style: render.Style{Bg: render.ColorBrightRed}, want: "\x1b[101mx\x1b[0m"},
		{name: "fundo RGB", style: render.Style{Bg: render.RGB(1, 2, 3)}, want: "\x1b[48;2;1;2;3mx\x1b[0m"},
		{name: "atributos e cor", style: render.Style{Bold: true, Italic: true, Fg: render.ColorGreen}, want: "\x1b[1;3;32mx\x1b[0m"},
		{name: "undercurl vira underline", style: render.Style{Undercurl: true}, want: "\x1b[4mx\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeSpan(&b, "x", tt.style)
			if got := b.String(); got != tt.want {
				t.Errorf("writeSpan = %q, esperava %q", got, tt.want)
			}
		})
	}
}

func TestRenderFrame(t *testing.T) {
	frame := render.Frame{
		Lines: []render.StyledLine{
			{{Text: "func", Style: render.Style{Fg: render.ColorMagenta}}, {Text: " main"}},
			{},
			{{Text: "~", Style: render.Style{Fg: render.ColorBrightBlack}}},
		},
		Status: render.StyledLine{{Text: " main.go"}},
	}
	want := "\x1b[35mfunc\x1b[0m main\n" +
		"\n" +
		"\x1b[90m~\x1b[0m\n" +
		" main.go"

	if got := RenderFrame(frame); got != want {
		t.Errorf("RenderFrame =\n%q\nesperava\n%q", got, want)
	}
}
