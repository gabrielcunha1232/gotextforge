package engine

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

func parseMarked(t *testing.T, marked string) (*buffer.Buffer, buffer.Pos) {
	t.Helper()
	var at buffer.Pos
	lines := strings.Split(marked, "\n")
	for i, l := range lines {
		if idx := strings.Index(l, "|"); idx >= 0 {
			at = buffer.Pos{Line: i, Col: len([]rune(l[:idx]))}
			lines[i] = l[:idx] + l[idx+1:]
		}
	}
	b, err := buffer.Load(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	return b, at
}

func mark(b *buffer.Buffer, p buffer.Pos) string {
	var out []string
	for i := 0; i < b.LineCount(); i++ {
		line := string(b.Line(i))
		if i == p.Line {
			r := []rune(line)
			c := min(p.Col, len(r))
			line = string(r[:c]) + "|" + string(r[c:])
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func TestMotions(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		motion Motion
		count  int
		op     bool
		want   string
		fail   bool
	}{
		{name: "w simples", src: "|foo bar", motion: Motion{Kind: MotionWordStart}, want: "foo |bar"},
		{name: "w para na pontuacao", src: "|foo.bar", motion: Motion{Kind: MotionWordStart}, want: "foo|.bar"},
		{name: "W pula a pontuacao", src: "|foo.bar baz", motion: Motion{Kind: MotionBigWordStart}, want: "foo.bar |baz"},
		{name: "3w", src: "|a b c d", motion: Motion{Kind: MotionWordStart}, count: 3, want: "a b c |d"},
		{name: "w atravessa linha", src: "fo|o\n  bar", motion: Motion{Kind: MotionWordStart}, want: "foo\n  |bar"},
		{name: "w para em linha vazia", src: "|foo\n\nbar", motion: Motion{Kind: MotionWordStart}, want: "foo\n|\nbar"},
		{name: "w no fim do arquivo fica no fim", src: "foo |bar", motion: Motion{Kind: MotionWordStart}, want: "foo bar|"},
		{name: "e", src: "|foo bar", motion: Motion{Kind: MotionWordEnd}, want: "fo|o bar"},
		{name: "e no fim da palavra vai para a proxima", src: "fo|o bar", motion: Motion{Kind: MotionWordEnd}, want: "foo ba|r"},
		{name: "e atravessa linha", src: "fo|o\n  bar", motion: Motion{Kind: MotionWordEnd}, want: "foo\n  ba|r"},
		{name: "b", src: "foo b|ar", motion: Motion{Kind: MotionWordBack}, want: "foo |bar"},
		{name: "b no inicio da palavra vai para a anterior", src: "foo |bar", motion: Motion{Kind: MotionWordBack}, want: "|foo bar"},
		{name: "b atravessa linha", src: "foo\n  |bar", motion: Motion{Kind: MotionWordBack}, want: "|foo\n  bar"},
		{name: "b para em linha vazia", src: "foo\n\n|bar", motion: Motion{Kind: MotionWordBack}, want: "foo\n|\nbar"},
		{name: "b no inicio do arquivo falha", src: "|foo", motion: Motion{Kind: MotionWordBack}, fail: true},
		{name: "B", src: "a.b c|d", motion: Motion{Kind: MotionBigWordBack}, want: "a.b |cd"},
		{name: "acentos sao letras", src: "|ação feita", motion: Motion{Kind: MotionWordStart}, want: "ação |feita"},
		{name: "0", src: "  fo|o", motion: Motion{Kind: MotionLineStart}, want: "|  foo"},
		{name: "^", src: "\t\tfo|o", motion: Motion{Kind: MotionFirstNonBlank}, want: "\t\t|foo"},
		{name: "$", src: "f|oo\nbar", motion: Motion{Kind: MotionLineEnd}, want: "fo|o\nbar"},
		{name: "2$ desce uma linha", src: "f|oo\nbarra", motion: Motion{Kind: MotionLineEnd}, count: 2, want: "foo\nbarr|a"},
		{name: "h", src: "fo|o", motion: Motion{Kind: MotionLeft}, want: "f|oo"},
		{name: "h na coluna 0 falha", src: "|foo", motion: Motion{Kind: MotionLeft}, fail: true},
		{name: "l no ultimo caractere falha", src: "fo|o", motion: Motion{Kind: MotionRight}, fail: true},
		{name: "l com operador vai ate depois do fim", src: "fo|o", motion: Motion{Kind: MotionRight}, op: true, want: "foo|"},
		{name: "j", src: "|a\nb\nc", motion: Motion{Kind: MotionDown}, want: "a\n|b\nc"},
		{name: "5j para na ultima", src: "|a\nb\nc", motion: Motion{Kind: MotionDown}, count: 5, want: "a\nb\n|c"},
		{name: "j na ultima linha falha", src: "a\n|b", motion: Motion{Kind: MotionDown}, fail: true},
		{name: "gg vai ao primeiro nao-branco", src: "  a\nb\n|c", motion: Motion{Kind: MotionFileStart}, want: "  |a\nb\nc"},
		{name: "G", src: "|a\nb\n\tc", motion: Motion{Kind: MotionFileEnd}, want: "a\nb\n\t|c"},
		{name: "3G vai para a linha 3", src: "|a\nb\nc\nd", motion: Motion{Kind: MotionFileEnd}, count: 3, want: "a\nb\n|c\nd"},
		{name: "fx", src: "|abcxdx", motion: Motion{Kind: MotionFindForward, Char: 'x'}, want: "abc|xdx"},
		{name: "2fx", src: "|abcxdx", motion: Motion{Kind: MotionFindForward, Char: 'x'}, count: 2, want: "abcxd|x"},
		{name: "tx", src: "|abcxd", motion: Motion{Kind: MotionTillForward, Char: 'x'}, want: "ab|cxd"},
		{name: "Fa", src: "abca|d", motion: Motion{Kind: MotionFindBackward, Char: 'a'}, want: "abc|ad"},
		{name: "Ta", src: "abcx|d", motion: Motion{Kind: MotionTillBackward, Char: 'a'}, want: "a|bcxd"},
		{name: "f sem achar falha", src: "|abc", motion: Motion{Kind: MotionFindForward, Char: 'z'}, fail: true},
		{name: "t colado nao anda", src: "|ax", motion: Motion{Kind: MotionTillForward, Char: 'x'}, fail: true},
		{name: "} vai para a linha vazia", src: "|a\nb\n\nc", motion: Motion{Kind: MotionParagraphForward}, want: "a\nb\n|\nc"},
		{name: "} na ultima linha vai para o fim", src: "|a\nb", motion: Motion{Kind: MotionParagraphForward}, want: "a\nb|"},
		{name: "{ volta para a linha vazia", src: "a\n\nb\n|c", motion: Motion{Kind: MotionParagraphBackward}, want: "a\n|\nb\nc"},
		{name: "2} pula dois paragrafos", src: "|a\n\nb\n\nc", motion: Motion{Kind: MotionParagraphForward}, count: 2, want: "a\n\nb\n|\nc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, at := parseMarked(t, tt.src)
			ctx := MotionContext{Count: tt.count, HasCount: tt.count > 0, Operator: tt.op}
			target, ok := ApplyMotion(b, at, tt.motion, ctx)
			if tt.fail {
				if ok {
					t.Errorf("deveria falhar, mas foi para %q", mark(b, target.Pos))
				}
				return
			}
			if !ok {
				t.Fatal("movimento falhou")
			}
			if got := mark(b, target.Pos); got != tt.want {
				t.Errorf("resultado %q\nesperava  %q", got, tt.want)
			}
		})
	}
}

func markRange(b *buffer.Buffer, r Range) string {
	start, end := r.Start, r.End
	text := b.Text(start, end)
	if r.Linewise {
		return "L:" + text
	}
	return text
}

func TestTextObjects(t *testing.T) {
	tests := []struct {
		name string
		src  string
		obj  ObjectKind
		in   bool
		want string
		fail bool
	}{
		{name: "iw", src: "foo b|ar baz", obj: ObjectWord, in: true, want: "bar"},
		{name: "aw inclui o espaco depois", src: "foo b|ar baz", obj: ObjectWord, want: "bar "},
		{name: "aw no fim usa o espaco antes", src: "foo b|ar", obj: ObjectWord, want: " bar"},
		{name: "iw em pontuacao", src: "a.|.b", obj: ObjectWord, in: true, want: ".."},
		{name: "iw em espaco", src: "a | b", obj: ObjectWord, in: true, want: "  "},
		{name: "iW", src: "x fo|o.bar y", obj: ObjectBigWord, in: true, want: "foo.bar"},
		{name: "iw em linha vazia falha", src: "a\n|\nb", obj: ObjectWord, in: true, fail: true},
		{name: "i\"", src: "x := \"ol|a mundo\"", obj: ObjectDoubleQuote, in: true, want: "ola mundo"},
		{name: "a\"", src: "f(\"ol|a\", 1)", obj: ObjectDoubleQuote, want: "\"ola\""},
		{name: "i\" com aspas escapadas", src: "s := \"a\\\"|b\"", obj: ObjectDoubleQuote, in: true, want: "a\\\"b"},
		{name: "i\" antes da string acha a proxima", src: "|x := \"abc\"", obj: ObjectDoubleQuote, in: true, want: "abc"},
		{name: "i\" vazio", src: "s := \"|\"", obj: ObjectDoubleQuote, in: true, want: ""},
		{name: "i\" sem aspas falha", src: "a|bc", obj: ObjectDoubleQuote, in: true, fail: true},
		{name: "i(", src: "f(a, |b)", obj: ObjectParen, in: true, want: "a, b"},
		{name: "a(", src: "f(a, |b)", obj: ObjectParen, want: "(a, b)"},
		{name: "i( aninhado", src: "f(a, g(|x), c)", obj: ObjectParen, in: true, want: "x"},
		{name: "i( de fora do aninhado", src: "f(a|, g(x), c)", obj: ObjectParen, in: true, want: "a, g(x), c"},
		{name: "i( com cursor no parentese", src: "f|(a)", obj: ObjectParen, in: true, want: "a"},
		{name: "i( com cursor no fecha", src: "f(a|)", obj: ObjectParen, in: true, want: "a"},
		{name: "i( sem parenteses falha", src: "a|bc", obj: ObjectParen, in: true, fail: true},
		{name: "i{ multilinha vira linhas inteiras", src: "func f() {\n\ta|()\n\tb()\n}", obj: ObjectBrace, in: true, want: "L:\ta()\n\tb()"},
		{name: "a{ multilinha", src: "if x {\n\t|y\n}", obj: ObjectBrace, want: "{\n\ty\n}"},
		{name: "i{ na mesma linha", src: "s := T{|a: 1}", obj: ObjectBrace, in: true, want: "a: 1"},
		{name: "ip", src: "a\n|b\n\nc", obj: ObjectParagraph, in: true, want: "L:a\nb"},
		{name: "ap inclui linhas vazias depois", src: "a\n|b\n\n\nc", obj: ObjectParagraph, want: "L:a\nb\n\n"},
		{name: "ap no ultimo paragrafo usa as vazias antes", src: "a\n\n|c\nd", obj: ObjectParagraph, want: "L:\nc\nd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, at := parseMarked(t, tt.src)
			r, ok := TextObject(b, at, Motion{Kind: MotionTextObject, Object: tt.obj, Inner: tt.in})
			if tt.fail {
				if ok {
					t.Errorf("deveria falhar, mas selecionou %q", markRange(b, r))
				}
				return
			}
			if !ok {
				t.Fatal("text object falhou")
			}
			if got := markRange(b, r); got != tt.want {
				t.Errorf("selecionou %q, esperava %q", got, tt.want)
			}
		})
	}
}
