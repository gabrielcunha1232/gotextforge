package engine

import (
	"testing"
)

func TestQueryMatches(t *testing.T) {
	tests := []struct {
		name string
		q    Query
		line string
		want []int
	}{
		{name: "simples", q: Query{Pattern: []rune("ab")}, line: "xabyab", want: []int{1, 4}},
		{name: "minusculas ignoram caixa", q: Query{Pattern: []rune("err")}, line: "Err err ERR", want: []int{0, 4, 8}},
		{name: "maiuscula torna sensivel", q: Query{Pattern: []rune("Err")}, line: "Err err ERR", want: []int{0}},
		{name: "acentos", q: Query{Pattern: []rune("ção")}, line: "AÇÃO ação", want: []int{1, 6}},
		{name: "sobrepostos", q: Query{Pattern: []rune("aa")}, line: "aaa", want: []int{0, 1}},
		{name: "palavra inteira", q: Query{Pattern: []rune("err"), WholeWord: true}, line: "err errs myerr err_x (err)", want: []int{0, 22}},
		{name: "padrao maior que a linha", q: Query{Pattern: []rune("abcdef")}, line: "abc", want: nil},
		{name: "padrao vazio", q: Query{}, line: "abc", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.q.Matches([]rune(tt.line))
			if len(got) != len(tt.want) {
				t.Fatalf("Matches = %v, esperava %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Matches = %v, esperava %v", got, tt.want)
				}
			}
		})
	}
}

func TestSearch(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		pattern     string
		forward     bool
		count       int
		want        string
		wantWrapped bool
		fail        bool
	}{
		{name: "proxima na mesma linha", src: "|foo bar foo", pattern: "foo", forward: true, want: "foo bar |foo"},
		{name: "proxima em outra linha", src: "|foo\nbar\nfoo", pattern: "foo", forward: true, want: "foo\nbar\n|foo"},
		{name: "volta ao inicio", src: "foo\nbar\n|foo", pattern: "foo", forward: true, want: "|foo\nbar\nfoo", wantWrapped: true},
		{name: "unica ocorrencia volta para ela mesma", src: "x |foo y", pattern: "foo", forward: true, want: "x |foo y", wantWrapped: true},
		{name: "para tras", src: "foo\nbar\n|foo", pattern: "foo", forward: false, want: "|foo\nbar\nfoo"},
		{name: "para tras na mesma linha", src: "foo bar |foo", pattern: "foo", forward: false, want: "|foo bar foo"},
		{name: "para tras volta ao fim", src: "|foo\nbar\nfoo", pattern: "foo", forward: false, want: "foo\nbar\n|foo", wantWrapped: true},
		{name: "count", src: "|a a a a", pattern: "a", forward: true, count: 3, want: "a a a |a"},
		{name: "nao encontrado", src: "|abc", pattern: "zz", forward: true, fail: true},
		{name: "ocorrencia antes do cursor na mesma linha com volta", src: "foo |bar\nx", pattern: "foo", forward: true, want: "|foo bar\nx", wantWrapped: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, at := parseMarked(t, tt.src)
			got, wrapped, ok := Search(b, at, Query{Pattern: []rune(tt.pattern)}, tt.forward, tt.count)
			if tt.fail {
				if ok {
					t.Errorf("deveria falhar, achou %q", mark(b, got))
				}
				return
			}
			if !ok {
				t.Fatal("busca falhou")
			}
			if m := mark(b, got); m != tt.want {
				t.Errorf("achou %q, esperava %q", m, tt.want)
			}
			if wrapped != tt.wantWrapped {
				t.Errorf("wrapped = %v, esperava %v", wrapped, tt.wantWrapped)
			}
		})
	}
}

func TestWordUnderCursor(t *testing.T) {
	tests := []struct {
		line  string
		col   int
		want  string
		start int
		ok    bool
	}{
		{"foo bar", 1, "foo", 0, true},
		{"foo bar", 3, "bar", 4, true},
		{"x.Println()", 4, "Println", 2, true},
		{"  ", 0, "", 0, false},
		{"", 0, "", 0, false},
	}
	for _, tt := range tests {
		got, start, ok := WordUnderCursor([]rune(tt.line), tt.col)
		if ok != tt.ok || string(got) != tt.want || (ok && start != tt.start) {
			t.Errorf("WordUnderCursor(%q, %d) = %q, %d, %v", tt.line, tt.col, string(got), start, ok)
		}
	}
}
