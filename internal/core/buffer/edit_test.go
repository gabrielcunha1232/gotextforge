package buffer

import (
	"strings"
	"testing"
)

func mustLoad(t testing.TB, src string) *Buffer {
	t.Helper()
	b, err := Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return b
}

func allText(b *Buffer) string {
	last := b.LineCount() - 1
	return b.Text(Pos{}, Pos{Line: last, Col: b.LineLen(last)})
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		at      Pos
		text    string
		want    string
		wantEnd Pos
	}{
		{name: "rune no meio", src: "abcd", at: Pos{0, 2}, text: "X", want: "abXcd", wantEnd: Pos{0, 3}},
		{name: "no inicio da linha", src: "abc", at: Pos{0, 0}, text: "X", want: "Xabc", wantEnd: Pos{0, 1}},
		{name: "no fim da linha", src: "abc", at: Pos{0, 3}, text: "X", want: "abcX", wantEnd: Pos{0, 4}},
		{name: "acentos contam como 1 rune", src: "ação", at: Pos{0, 2}, text: "é", want: "açéão", wantEnd: Pos{0, 3}},
		{name: "buffer vazio", src: "", at: Pos{0, 0}, text: "package main", want: "package main", wantEnd: Pos{0, 12}},
		{name: "enter no meio quebra a linha", src: "abcd", at: Pos{0, 2}, text: "\n", want: "ab\ncd", wantEnd: Pos{1, 0}},
		{name: "enter no fim cria linha vazia abaixo", src: "abc", at: Pos{0, 3}, text: "\n", want: "abc\n", wantEnd: Pos{1, 0}},
		{name: "enter no inicio cria linha vazia acima", src: "abc", at: Pos{0, 0}, text: "\n", want: "\nabc", wantEnd: Pos{1, 0}},
		{name: "texto com varias linhas", src: "ab\nzz", at: Pos{0, 1}, text: "1\n2\n3", want: "a1\n2\n3b\nzz", wantEnd: Pos{2, 1}},
		{name: "CRLF no texto vira quebra normal", src: "ab", at: Pos{0, 1}, text: "x\r\ny", want: "ax\nyb", wantEnd: Pos{1, 1}},
		{name: "na segunda linha", src: "um\ndois\ntres", at: Pos{1, 4}, text: "!", want: "um\ndois!\ntres", wantEnd: Pos{1, 5}},
		{name: "texto vazio nao muda nada", src: "abc", at: Pos{0, 1}, text: "", want: "abc", wantEnd: Pos{0, 1}},
		{name: "posicao alem do fim da linha e limitada", src: "abc", at: Pos{0, 99}, text: "X", want: "abcX", wantEnd: Pos{0, 4}},
		{name: "posicao alem da ultima linha e limitada", src: "abc", at: Pos{9, 0}, text: "X", want: "Xabc", wantEnd: Pos{0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustLoad(t, tt.src)
			end := b.Insert(tt.at, tt.text)
			if got := allText(b); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if end != tt.wantEnd {
				t.Errorf("Insert devolveu %+v, esperava %+v", end, tt.wantEnd)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		start, end  Pos
		want        string
		wantRemoved string
	}{
		{name: "um caractere", src: "abcd", start: Pos{0, 1}, end: Pos{0, 2}, want: "acd", wantRemoved: "b"},
		{name: "linha inteira sem a quebra", src: "abc\nxyz", start: Pos{0, 0}, end: Pos{0, 3}, want: "\nxyz", wantRemoved: "abc"},
		{name: "juntar linhas (backspace no inicio)", src: "abc\nxyz", start: Pos{0, 3}, end: Pos{1, 0}, want: "abcxyz", wantRemoved: "\n"},
		{name: "varias linhas", src: "um\ndois\ntres\nquatro", start: Pos{0, 1}, end: Pos{2, 2}, want: "ues\nquatro", wantRemoved: "m\ndois\ntr"},
		{name: "inicio e fim invertidos", src: "abcd", start: Pos{0, 3}, end: Pos{0, 1}, want: "ad", wantRemoved: "bc"},
		{name: "intervalo vazio", src: "abcd", start: Pos{0, 2}, end: Pos{0, 2}, want: "abcd", wantRemoved: ""},
		{name: "tudo vira uma linha vazia", src: "a\nb\nc", start: Pos{0, 0}, end: Pos{2, 1}, want: "", wantRemoved: "a\nb\nc"},
		{name: "acentos", src: "ação", start: Pos{0, 1}, end: Pos{0, 3}, want: "ao", wantRemoved: "çã"},
		{name: "fim alem da linha e limitado", src: "abc\nxyz", start: Pos{0, 1}, end: Pos{0, 99}, want: "a\nxyz", wantRemoved: "bc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mustLoad(t, tt.src)
			removed := b.Delete(tt.start, tt.end)
			if got := allText(b); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if removed != tt.wantRemoved {
				t.Errorf("Delete devolveu %q, esperava %q", removed, tt.wantRemoved)
			}
			if b.LineCount() < 1 {
				t.Errorf("buffer ficou com %d linhas; precisa ter pelo menos 1", b.LineCount())
			}
		})
	}
}

func TestInsertDeleteIdaEVolta(t *testing.T) {
	const src = "package main\n\nfunc main() {\n\tfmt.Println(\"ação\")\n}\n"
	cases := []struct {
		at   Pos
		text string
	}{
		{Pos{0, 7}, "X"},
		{Pos{3, 1}, "\n"},
		{Pos{2, 0}, "// novo\n// bloco\n"},
		{Pos{3, 14}, "é\nç"},
	}

	for _, c := range cases {
		b := mustLoad(t, src)
		before := allText(b)
		end := b.Insert(c.at, c.text)
		removed := b.Delete(c.at, end)
		if removed != c.text {
			t.Errorf("Insert(%+v, %q) e Delete devolveu %q", c.at, c.text, removed)
		}
		if got := allText(b); got != before {
			t.Errorf("Insert(%+v, %q) e Delete: texto = %q, esperava o original %q", c.at, c.text, got, before)
		}
	}

	ranges := [][2]Pos{
		{{0, 0}, {0, 7}},
		{{2, 5}, {4, 1}},
		{{1, 0}, {2, 0}},
	}
	for _, r := range ranges {
		b := mustLoad(t, src)
		before := allText(b)
		removed := b.Delete(r[0], r[1])
		b.Insert(r[0], removed)
		if got := allText(b); got != before {
			t.Errorf("Delete(%+v) e Insert: texto = %q, esperava o original %q", r, got, before)
		}
	}
}

func TestModified(t *testing.T) {
	b := mustLoad(t, "abc")
	if b.Modified() {
		t.Fatal("buffer recem carregado nao deveria estar modificado")
	}
	b.Insert(Pos{0, 0}, "")
	if b.Modified() {
		t.Fatal("inserir texto vazio nao deveria marcar como modificado")
	}
	b.Insert(Pos{0, 0}, "x")
	if !b.Modified() {
		t.Fatal("depois de Insert deveria estar modificado")
	}
	b.MarkSaved()
	if b.Modified() {
		t.Fatal("depois de MarkSaved nao deveria estar modificado")
	}
	b.Delete(Pos{0, 0}, Pos{0, 0})
	if b.Modified() {
		t.Fatal("Delete de intervalo vazio nao deveria marcar como modificado")
	}
	b.Delete(Pos{0, 0}, Pos{0, 1})
	if !b.Modified() {
		t.Fatal("depois de Delete deveria estar modificado")
	}
}

func TestInsertNaoAlteraCopiaDeLine(t *testing.T) {
	b := mustLoad(t, "abcdef")
	snapshot := b.Line(0)
	b.Insert(Pos{0, 3}, "\n")
	if string(snapshot) != "abcdef" {
		t.Errorf("copia obtida por Line() mudou para %q", string(snapshot))
	}
}
