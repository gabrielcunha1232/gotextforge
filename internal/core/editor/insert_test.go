package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

var specialKeys = map[string]input.KeyCode{
	"<Esc>":   input.KeyEsc,
	"<CR>":    input.KeyEnter,
	"<BS>":    input.KeyBackspace,
	"<Del>":   input.KeyDelete,
	"<Tab>":   input.KeyTab,
	"<Up>":    input.KeyUp,
	"<Down>":  input.KeyDown,
	"<Left>":  input.KeyLeft,
	"<Right>": input.KeyRight,
	"<End>":   input.KeyEnd,
	"<Home>":  input.KeyHome,
}

func typeKeys(t *testing.T, ed *Editor, keys string) {
	t.Helper()
	for len(keys) > 0 {
		if keys[0] == '<' {
			end := strings.IndexByte(keys, '>')
			if end < 0 {
				t.Fatalf("tecla especial sem '>' em %q", keys)
			}
			name := keys[:end+1]
			code, ok := specialKeys[name]
			if !ok {
				t.Fatalf("tecla especial desconhecida %q", name)
			}
			ed.HandleEvent(input.NewKeyEvent(input.Key{Code: code}))
			keys = keys[end+1:]
			continue
		}
		r := []rune(keys)[0]
		ed.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyRune, Rune: r}))
		keys = keys[len(string(r)):]
	}
}

func bufferText(ed *Editor) string {
	last := ed.buf.LineCount() - 1
	return ed.buf.Text(buffer.Pos{}, buffer.Pos{Line: last, Col: ed.buf.LineLen(last)})
}

func TestInsertMode(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		keys       string
		want       string
		wantCursor buffer.Pos
		wantMode   Mode
	}{
		{name: "i entra no insert e digita antes do cursor", src: "abc", keys: "lXY", want: "abc", wantCursor: pos(0, 1), wantMode: ModeNormal},
		{name: "i digita na posicao do cursor", src: "abc", keys: "liXY", want: "aXYbc", wantCursor: pos(0, 3), wantMode: ModeInsert},
		{name: "Esc volta ao normal e o cursor anda 1 para a esquerda", src: "abc", keys: "liXY<Esc>", want: "aXYbc", wantCursor: pos(0, 2), wantMode: ModeNormal},
		{name: "Esc na coluna 0 nao anda", src: "abc", keys: "i<Esc>", want: "abc", wantCursor: pos(0, 0), wantMode: ModeNormal},
		{name: "a digita depois do cursor", src: "abc", keys: "aX", want: "aXbc", wantCursor: pos(0, 2), wantMode: ModeInsert},
		{name: "A digita no fim da linha", src: "abc", keys: "AX", want: "abcX", wantCursor: pos(0, 4), wantMode: ModeInsert},
		{name: "I digita no inicio da linha", src: "abc", keys: "$IX", want: "Xabc", wantCursor: pos(0, 1), wantMode: ModeInsert},
		{name: "a em linha vazia", src: "", keys: "aX", want: "X", wantCursor: pos(0, 1), wantMode: ModeInsert},
		{name: "o abre linha abaixo", src: "abc\nxyz", keys: "oNOVA", want: "abc\nNOVA\nxyz", wantCursor: pos(1, 4), wantMode: ModeInsert},
		{name: "O abre linha acima", src: "abc\nxyz", keys: "jONOVA", want: "abc\nNOVA\nxyz", wantCursor: pos(1, 4), wantMode: ModeInsert},
		{name: "Enter no meio da linha", src: "abcd", keys: "lli<CR>", want: "ab\ncd", wantCursor: pos(1, 0), wantMode: ModeInsert},
		{name: "Tab insere tab", src: "x", keys: "i<Tab>", want: "\tx", wantCursor: pos(0, 1), wantMode: ModeInsert},
		{name: "espaco e acentos", src: "", keys: "ifunção é", want: "função é", wantCursor: pos(0, 8), wantMode: ModeInsert},
		{name: "Backspace apaga o anterior", src: "abc", keys: "A<BS>", want: "ab", wantCursor: pos(0, 2), wantMode: ModeInsert},
		{name: "Backspace no inicio junta com a linha de cima", src: "abc\nxyz", keys: "ji<BS>", want: "abcxyz", wantCursor: pos(0, 3), wantMode: ModeInsert},
		{name: "Backspace no inicio do arquivo nao faz nada", src: "abc", keys: "i<BS>", want: "abc", wantCursor: pos(0, 0), wantMode: ModeInsert},
		{name: "Del apaga o caractere sob o cursor", src: "abc", keys: "i<Del>", want: "bc", wantCursor: pos(0, 0), wantMode: ModeInsert},
		{name: "Del no fim da linha junta com a de baixo", src: "abc\nxyz", keys: "A<Del>", want: "abcxyz", wantCursor: pos(0, 3), wantMode: ModeInsert},
		{name: "Del no fim do arquivo nao faz nada", src: "abc", keys: "A<Del>", want: "abc", wantCursor: pos(0, 3), wantMode: ModeInsert},
		{name: "no insert hjkl sao texto, nao movimento", src: "", keys: "ihjkl", want: "hjkl", wantCursor: pos(0, 4), wantMode: ModeInsert},
		{name: "setas movem no insert", src: "abc\nxyz", keys: "A<Down>", want: "abc\nxyz", wantCursor: pos(1, 3), wantMode: ModeInsert},
		{name: "no insert o cursor pode ficar depois do ultimo caractere", src: "abc", keys: "i<End>", want: "abc", wantCursor: pos(0, 3), wantMode: ModeInsert},
		{name: "x apaga sob o cursor no normal", src: "abc", keys: "lx", want: "ac", wantCursor: pos(0, 1), wantMode: ModeNormal},
		{name: "x no ultimo caractere recua o cursor", src: "abc", keys: "$x", want: "ab", wantCursor: pos(0, 1), wantMode: ModeNormal},
		{name: "x em linha vazia nao junta linhas", src: "\nabc", keys: "x", want: "\nabc", wantCursor: pos(0, 0), wantMode: ModeNormal},
		{name: "fluxo completo", src: "package main\n", keys: "o<CR>func main() {<CR>}<Esc>kA<CR><Tab>println(1)<Esc>", want: "package main\n\nfunc main() {\n\tprintln(1)\n}", wantCursor: pos(3, 10), wantMode: ModeNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, tt.src)
			typeKeys(t, ed, tt.keys)
			if got := bufferText(ed); got != tt.want {
				t.Errorf("texto = %q, esperava %q", got, tt.want)
			}
			if got := ed.Cursor(); got != tt.wantCursor {
				t.Errorf("Cursor() = %+v, esperava %+v", got, tt.wantCursor)
			}
			if got := ed.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, esperava %v", got, tt.wantMode)
			}
		})
	}
}

func TestInsertModeFrame(t *testing.T) {
	ed := editorFrom(t, "abc\n")

	frame := ed.Frame()
	if frame.CursorShape != render.CursorBlock {
		t.Errorf("no normal o cursor deveria ser bloco, veio %v", frame.CursorShape)
	}
	if strings.Contains(frame.Status, "INSERT") || strings.Contains(frame.Status, "[+]") {
		t.Errorf("status inicial nao deveria ter INSERT nem [+]: %q", frame.Status)
	}

	typeKeys(t, ed, "AX")
	frame = ed.Frame()
	if frame.CursorShape != render.CursorBar {
		t.Errorf("no insert o cursor deveria ser barra, veio %v", frame.CursorShape)
	}
	if !strings.Contains(frame.Status, "-- INSERT --") {
		t.Errorf("status deveria mostrar -- INSERT --: %q", frame.Status)
	}
	if !strings.Contains(frame.Status, "[+]") {
		t.Errorf("status deveria mostrar [+] depois de editar: %q", frame.Status)
	}
	if frame.Cursor != (render.ScreenPos{Row: 0, Col: 4}) {
		t.Errorf("cursor depois do ultimo caractere deveria estar na coluna 4, veio %+v", frame.Cursor)
	}
	if got := frameLineText(frame.Lines[0]); got != "abcX" {
		t.Errorf("linha desenhada = %q, esperava %q", got, "abcX")
	}

	typeKeys(t, ed, "<Esc>")
	if strings.Contains(ed.Frame().Status, "INSERT") {
		t.Errorf("depois de Esc o status nao deveria ter INSERT: %q", ed.Frame().Status)
	}
}

func TestDigitarNoFimDaTelaRola(t *testing.T) {
	buf := makeNumberedBuffer(t, 3)
	ed := New(buf, "teste.go")
	ed.HandleEvent(input.NewResizeEvent(80, 4))

	typeKeys(t, ed, "GA<CR>4<CR>5<CR>6")

	frame := ed.Frame()
	if ed.Cursor().Line != 5 {
		t.Fatalf("cursor na linha %d, esperava 5", ed.Cursor().Line)
	}
	if frame.Cursor.Row != 2 {
		t.Errorf("cursor deveria estar no rodape (linha 2 da tela), veio %d", frame.Cursor.Row)
	}
	if got := frameLineText(frame.Lines[2]); got != "6" {
		t.Errorf("ultima linha visivel = %q, esperava %q", got, "6")
	}
}

func pos(line, col int) buffer.Pos {
	return buffer.Pos{Line: line, Col: col}
}
