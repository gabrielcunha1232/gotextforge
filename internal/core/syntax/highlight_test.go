package syntax

import (
	"fmt"
	"strings"
	"testing"
)

var kindNames = map[Kind]string{
	KindKeyword:     "kw",
	KindString:      "str",
	KindNumber:      "num",
	KindComment:     "com",
	KindPredeclared: "pre",
	KindFunc:        "fn",
}

func describe(line string, tokens []Token) string {
	runes := []rune(line)
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		parts = append(parts, fmt.Sprintf("%s(%s)", kindNames[t.Kind], string(runes[t.Start:t.End])))
	}
	return strings.Join(parts, " ")
}

func TestHighlightLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		in      State
		want    string
		wantOut State
	}{
		{name: "linha vazia", line: "", want: ""},
		{name: "package", line: "package main", want: "kw(package)"},
		{name: "func com nome", line: "func main() {", want: "kw(func) fn(main)"},
		{name: "chamada com string", line: "\tfmt.Println(\"oi\")", want: "fn(Println) str(\"oi\")"},
		{name: "tipos e valores predeclarados", line: "var x int = len(s) + 42", want: "kw(var) pre(int) pre(len) num(42)"},
		{name: "nil e erro", line: "if err != nil { return err }", want: "kw(if) pre(nil) kw(return)"},
		{name: "numeros", line: "a := 0x1F + 3.14 + 2i + 1_000", want: "num(0x1F) num(3.14) num(2i) num(1_000)"},
		{name: "rune literal", line: "c := 'x'", want: "str('x')"},
		{name: "comentario de linha", line: "x := 1 // nota: return", want: "num(1) com(// nota: return)"},
		{name: "palavra-chave dentro de string nao conta", line: "s := \"func return\"", want: "str(\"func return\")"},
		{name: "acentos no meio usam indices de rune", line: "s := \"ação\" + len(t)", want: "str(\"ação\") pre(len)"},
		{name: "comentario de bloco na mesma linha", line: "a /* x */ := 1", want: "com(/* x */) num(1)"},
		{name: "comentario de bloco que nao termina", line: "x := 1 /* comeca aqui", want: "num(1) com(/* comeca aqui)", wantOut: StateBlockComment},
		{name: "dentro de comentario de bloco", line: "ainda func comentario", in: StateBlockComment, want: "com(ainda func comentario)", wantOut: StateBlockComment},
		{name: "fim de comentario de bloco", line: "fim */ return 1", in: StateBlockComment, want: "com(fim */) kw(return) num(1)"},
		{name: "raw string que nao termina", line: "q := `select *", want: "str(`select *)", wantOut: StateRawString},
		{name: "dentro de raw string", line: "from tabela where func = 1", in: StateRawString, want: "str(from tabela where func = 1)", wantOut: StateRawString},
		{name: "fim de raw string", line: "where id = ?` // sql", in: StateRawString, want: "str(where id = ?`) com(// sql)"},
		{name: "raw string completa", line: "s := `a\"b`", want: "str(`a\"b`)"},
		{name: "abre e fecha bloco e abre de novo", line: "/* a */ x /* b", want: "com(/* a */) com(/* b)", wantOut: StateBlockComment},
		{name: "/*/ nao fecha o comentario", line: "/*/", want: "com(/*/)", wantOut: StateBlockComment},
		{name: "metodo com receiver", line: "func (e *Editor) Frame() render.Frame {", want: "kw(func) fn(Frame)"},
		{name: "generics", line: "func Map[T any](xs []T) {}", want: "kw(func) pre(any)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, out := HighlightLine([]rune(tt.line), tt.in)
			if got := describe(tt.line, tokens); got != tt.want {
				t.Errorf("tokens = %s\nesperava   %s", got, tt.want)
			}
			if out != tt.wantOut {
				t.Errorf("estado final = %v, esperava %v", out, tt.wantOut)
			}
		})
	}
}

func TestHighlightTokensDentroDaLinha(t *testing.T) {
	lines := []string{
		"package main",
		"\tfmt.Println(\"ação 世界\") // comentário",
		"x := `raw",
		"a /* b */ c",
		"\x00\x01 caracteres estranhos ☃ \"sem fim",
	}
	for _, line := range lines {
		for _, in := range []State{StateNormal, StateBlockComment, StateRawString} {
			tokens, _ := HighlightLine([]rune(line), in)
			n := len([]rune(line))
			prev := 0
			for _, tk := range tokens {
				if tk.Start < prev || tk.End <= tk.Start || tk.End > n {
					t.Errorf("linha %q estado %v: token invalido %+v (anterior terminou em %d, linha tem %d runes)", line, in, tk, prev, n)
				}
				prev = tk.End
			}
		}
	}
}
