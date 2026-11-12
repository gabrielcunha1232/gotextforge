package view

import "testing"

func TestFormatLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		width   int
		tabStop int
		want    string
	}{
		{
			name:    "linha vazia",
			line:    "",
			width:   80,
			tabStop: 4,
			want:    "",
		},
		{
			name:    "tab no inicio da linha expande para 4 espacos",
			line:    "\tfunc main()",
			width:   80,
			tabStop: 4,
			want:    "    func main()",
		},
		{
			name:    "tab no meio da linha alinha na proxima parada de 4",
			line:    "ab\tcd",
			width:   80,
			tabStop: 4,
			want:    "ab  cd",
		},
		{
			name:    "linha comprida cortada na largura da tela",
			line:    "abcdefghij",
			width:   5,
			tabStop: 4,
			want:    "abcde",
		},
		{
			name:    "tab cortada no meio pela largura da tela",
			line:    "ab\tcd",
			width:   3,
			tabStop: 4,
			want:    "ab ",
		},
		{
			name:    "caracteres acentuados cortados por rune e nao por byte",
			line:    "ação rápida",
			width:   4,
			tabStop: 4,
			want:    "ação",
		},
		{
			name:    "caractere CR solto vira ^M",
			line:    "a\rb",
			width:   80,
			tabStop: 4,
			want:    "a^Mb",
		},
		{
			name:    "sequencia ANSI ESC vira ^[ e nao executa comando no terminal",
			line:    "\x1b[2Jlimpar",
			width:   80,
			tabStop: 4,
			want:    "^[[2Jlimpar",
		},
		{
			name:    "caractere DEL (0x7f) vira ^?",
			line:    "abc\x7fdef",
			width:   80,
			tabStop: 4,
			want:    "abc^?def",
		},
		{
			name:    "notacao de controle cortada no meio pela largura da tela",
			line:    "a\rb",
			width:   2,
			tabStop: 4,
			want:    "a^",
		},
		{
			name:    "controle C1 (CSI 0x9b) vira <9b>",
			line:    "a\u009b2Jb",
			width:   80,
			tabStop: 4,
			want:    "a<9b>2Jb",
		},
		{
			name:    "controle C1 (0x80) vira <80>",
			line:    "\u0080x",
			width:   80,
			tabStop: 4,
			want:    "<80>x",
		},
		{
			name:    "notacao C1 cortada no meio pela largura da tela",
			line:    "a\u009bb",
			width:   3,
			tabStop: 4,
			want:    "a<9",
		},
		{
			name:    "acentos latinos acima de 0x9f nao sao controle",
			line:    " é",
			width:   80,
			tabStop: 4,
			want:    " é",
		},
		{
			name:    "largura zero retorna vazio",
			line:    "texto",
			width:   0,
			tabStop: 4,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatLine([]rune(tt.line), tt.width, tt.tabStop)
			if got != tt.want {
				t.Errorf("FormatLine(%q, %d, %d) = %q, esperava %q",
					tt.line, tt.width, tt.tabStop, got, tt.want)
			}
		})
	}
}
