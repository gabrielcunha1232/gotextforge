package buffer

import (
	"strings"
	"testing"
)

func TestBufferLoad(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantLines      []string
		wantLineEnding LineEnding
	}{
		{
			name:           "arquivo vazio tem 1 linha vazia",
			input:          "",
			wantLines:      []string{""},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "apenas uma quebra de linha",
			input:          "\n",
			wantLines:      []string{""},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "duas quebras de linha preservam linha em branco",
			input:          "\n\n",
			wantLines:      []string{"", ""},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "LF simples sem newline final",
			input:          "linha 1\nlinha 2",
			wantLines:      []string{"linha 1", "linha 2"},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "LF com newline final nao cria linha extra",
			input:          "a\nb\n",
			wantLines:      []string{"a", "b"},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "LF com linha em branco no fim preserva linha vazia",
			input:          "a\n\n",
			wantLines:      []string{"a", ""},
			wantLineEnding: LineEndingLF,
		},
		{
			name:           "CRLF do Windows remove \\r e detecta terminador",
			input:          "a\r\nb\r\n",
			wantLines:      []string{"a", "b"},
			wantLineEnding: LineEndingCRLF,
		},
		{
			name:           "CRLF com linha em branco no fim preserva linha vazia",
			input:          "a\r\n\r\n",
			wantLines:      []string{"a", ""},
			wantLineEnding: LineEndingCRLF,
		},
		{
			name:           "caracteres UTF-8 acentuados e CJK contam em runes",
			input:          "ação\n世界",
			wantLines:      []string{"ação", "世界"},
			wantLineEnding: LineEndingLF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := Load(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Load falhou inesperadamente: %v", err)
			}

			if buf.LineCount() != len(tt.wantLines) {
				t.Fatalf("LineCount() = %d, esperava %d", buf.LineCount(), len(tt.wantLines))
			}

			if buf.LineEnding() != tt.wantLineEnding {
				t.Errorf("LineEnding() = %v, esperava %v", buf.LineEnding(), tt.wantLineEnding)
			}

			for i, want := range tt.wantLines {
				gotRunes := buf.Line(i)
				if got := string(gotRunes); got != want {
					t.Errorf("Line(%d) = %q, esperava %q", i, got, want)
				}
				if len(gotRunes) != len([]rune(want)) {
					t.Errorf("len(Line(%d)) = %d runes, esperava %d", i, len(gotRunes), len([]rune(want)))
				}
			}

			if buf.Line(-1) != nil || buf.Line(buf.LineCount()) != nil {
				t.Errorf("esperava nil ao acessar índices fora dos limites")
			}
		})
	}
}
