package buffer

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func writeString(t *testing.T, b *Buffer) string {
	t.Helper()
	var out bytes.Buffer
	n, err := b.WriteTo(&out)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n != int64(out.Len()) {
		t.Errorf("WriteTo devolveu n=%d, mas escreveu %d bytes", n, out.Len())
	}
	return out.String()
}

func TestWriteToIdaEVolta(t *testing.T) {
	inputs := map[string]string{
		"vazio":                         "",
		"so um newline":                 "\n",
		"sem newline final":             "abc",
		"com newline final":             "abc\n",
		"linhas em branco no fim":       "a\n\n\n",
		"linhas em branco no inicio":    "\n\na",
		"CRLF":                          "a\r\nb\r\n",
		"CRLF sem newline final":        "a\r\nb",
		"BOM UTF-8":                     "\xef\xbb\xbfpackage main\n",
		"BOM com CRLF":                  "\xef\xbb\xbfa\r\nb\r\n",
		"codigo Go real":                "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"ação\")\n}\n",
		"tabs e acentos sem newline":    "\tção\t世界",
		"linha so com espacos no final": "a\n   ",
	}

	for name, src := range inputs {
		t.Run(name, func(t *testing.T) {
			b := mustLoad(t, src)
			if got := writeString(t, b); got != src {
				t.Errorf("Load+WriteTo = %q, esperava o original %q", got, src)
			}
		})
	}
}

func TestWriteToDepoisDeEditar(t *testing.T) {
	tests := []struct {
		name string
		src  string
		edit func(b *Buffer)
		want string
	}{
		{
			name: "CRLF continua CRLF nas linhas novas",
			src:  "a\r\nb\r\n",
			edit: func(b *Buffer) { b.Insert(Pos{Line: 0, Col: 1}, "\nnova") },
			want: "a\r\nnova\r\nb\r\n",
		},
		{
			name: "BOM continua depois de editar a primeira linha",
			src:  "\xef\xbb\xbfabc\n",
			edit: func(b *Buffer) { b.Insert(Pos{Line: 0, Col: 0}, "X") },
			want: "\xef\xbb\xbfXabc\n",
		},
		{
			name: "arquivo sem newline final continua sem",
			src:  "abc",
			edit: func(b *Buffer) { b.Insert(Pos{Line: 0, Col: 3}, "\ndef") },
			want: "abc\ndef",
		},
		{
			name: "apagar tudo de um arquivo com newline final",
			src:  "a\nb\n",
			edit: func(b *Buffer) { b.Delete(Pos{Line: 0, Col: 0}, Pos{Line: 1, Col: 1}) },
			want: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustLoad(t, tt.src)
			tt.edit(b)
			if got := writeString(t, b); got != tt.want {
				t.Errorf("WriteTo = %q, esperava %q", got, tt.want)
			}
		})
	}
}

func TestWriteToBufferNovo(t *testing.T) {
	b := New()
	b.Insert(Pos{}, "package main")
	if got := writeString(t, b); got != "package main\n" {
		t.Errorf("buffer novo deveria terminar com newline, veio %q", got)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disco cheio") }

func TestWriteToPropagaErro(t *testing.T) {
	b := mustLoad(t, strings.Repeat("linha comprida de teste\n", 1000))
	if _, err := b.WriteTo(failingWriter{}); err == nil {
		t.Fatal("esperava erro do writer, veio nil")
	}
}

func TestHasInvalidUTF8(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want bool
	}{
		{name: "ASCII", src: "abc\n", want: false},
		{name: "acentos validos", src: "ação\n", want: false},
		{name: "byte invalido", src: "a\xffb\n", want: true},
		{name: "Latin-1 (nao UTF-8)", src: "a\xe7\xe3o\n", want: true},
		{name: "vazio", src: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustLoad(t, tt.src).HasInvalidUTF8(); got != tt.want {
				t.Errorf("HasInvalidUTF8() = %v, esperava %v", got, tt.want)
			}
		})
	}
}

func TestPlainTextEVersion(t *testing.T) {
	b := mustLoad(t, "\xef\xbb\xbfa\r\nb\r\n")
	if got := b.PlainText(); got != "a\nb\n" {
		t.Errorf("PlainText = %q, esperava sem BOM e com LF", got)
	}
	v1 := b.Version()
	b.Insert(Pos{}, "x")
	v2 := b.Version()
	b.Undo()
	v3 := b.Version()
	if !(v1 < v2 && v2 < v3) {
		t.Errorf("Version deveria crescer a cada mudanca, inclusive undo: %d %d %d", v1, v2, v3)
	}
}
