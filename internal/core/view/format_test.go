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
