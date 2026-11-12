package editor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func frameLineText(line render.StyledLine) string {
	spans := line
	if len(spans) > 0 && (spans[0].Style == styleLineNr || spans[0].Style == styleCursorLineNr) {
		spans = spans[1:]
	}
	if n := len(spans); n > 0 && spans[n-1].Style == (render.Style{Bg: colBgHighlight}) && strings.TrimSpace(spans[n-1].Text) == "" {
		spans = spans[:n-1]
	}
	var b strings.Builder
	for _, span := range spans {
		b.WriteString(span.Text)
	}
	return b.String()
}

func makeNumberedBuffer(t *testing.T, count int) *buffer.Buffer {
	t.Helper()
	var b strings.Builder
	for i := 1; i <= count; i++ {
		fmt.Fprintf(&b, "linha %d\n", i)
	}
	buf, err := buffer.Load(strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("erro ao criar buffer de teste: %v", err)
	}
	return buf
}

func TestEditorViewportAndScroll(t *testing.T) {
	t.Run("100 linhas com altura 11 produzem Frame com 10 linhas visiveis", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 100)
		ed := New(buf, "teste.go")
		ed.HandleEvent(input.NewResizeEvent(80, 11))

		frame := ed.Frame()
		if len(frame.Lines) != 10 {
			t.Fatalf("len(frame.Lines) = %d, esperava 10", len(frame.Lines))
		}
		if got := frameLineText(frame.Lines[0]); got != "linha 1" {
			t.Errorf("primeira linha = %q, esperava %q", got, "linha 1")
		}
		if got := frameLineText(frame.Lines[9]); got != "linha 10" {
			t.Errorf("decima linha = %q, esperava %q", got, "linha 10")
		}
	})

	t.Run("tecla j move o cursor sem rolar enquanto ele esta visivel", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 100)
		ed := New(buf, "teste.go")
		ed.HandleEvent(input.NewResizeEvent(80, 11))

		press(ed, 'j')

		frame := ed.Frame()
		if got := frameLineText(frame.Lines[0]); got != "linha 1" {
			t.Errorf("primeira linha apos 'j' = %q, esperava %q", got, "linha 1")
		}
		if frame.Cursor != (render.ScreenPos{Row: 1, Col: ed.gutterWidth()}) {
			t.Errorf("Cursor = %+v, esperava {Row:1 Col:%d}", frame.Cursor, ed.gutterWidth())
		}
	})

	t.Run("cursor passando do rodape rola a tela", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 100)
		ed := New(buf, "teste.go")
		ed.HandleEvent(input.NewResizeEvent(80, 11))

		for i := 0; i < 10; i++ {
			press(ed, 'j')
		}

		frame := ed.Frame()
		if got := frameLineText(frame.Lines[0]); got != "linha 2" {
			t.Errorf("primeira linha = %q, esperava %q", got, "linha 2")
		}
		if frame.Cursor.Row != 9 {
			t.Errorf("Cursor.Row = %d, esperava 9 (rodape)", frame.Cursor.Row)
		}
	})

	t.Run("tecla G vai para o fim com ultima linha na base", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 100)
		ed := New(buf, "teste.go")
		ed.HandleEvent(input.NewResizeEvent(80, 11))

		ed.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyRune, Rune: 'G'}))

		frame := ed.Frame()
		if got := frameLineText(frame.Lines[0]); got != "linha 91" {
			t.Errorf("primeira linha apos 'G' = %q, esperava %q", got, "linha 91")
		}
		if got := frameLineText(frame.Lines[9]); got != "linha 100" {
			t.Errorf("ultima linha apos 'G' = %q, esperava %q", got, "linha 100")
		}
	})

	t.Run("arquivo curto e completado com til (~)", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 3)
		ed := New(buf, "curto.go")
		ed.HandleEvent(input.NewResizeEvent(80, 6))

		frame := ed.Frame()
		if len(frame.Lines) != 5 {
			t.Fatalf("len(frame.Lines) = %d, esperava 5", len(frame.Lines))
		}
		if got := frameLineText(frame.Lines[2]); got != "linha 3" {
			t.Errorf("linha 3 = %q, esperava %q", got, "linha 3")
		}
		if got := frameLineText(frame.Lines[3]); got != "~" {
			t.Errorf("linha 4 (padding) = %q, esperava %q", got, "~")
		}
		if got := frameLineText(frame.Lines[4]); got != "~" {
			t.Errorf("linha 5 (padding) = %q, esperava %q", got, "~")
		}
	})
}

func press(ed *Editor, r rune) {
	ed.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyRune, Rune: r}))
}

func pressCode(ed *Editor, c input.KeyCode) {
	ed.HandleEvent(input.NewKeyEvent(input.Key{Code: c}))
}

func editorFrom(t *testing.T, src string) *Editor {
	t.Helper()
	buf, err := buffer.Load(strings.NewReader(src))
	if err != nil {
		t.Fatalf("erro ao criar buffer de teste: %v", err)
	}
	ed := New(buf, "teste.go")
	ed.HandleEvent(input.NewResizeEvent(80, 24))
	return ed
}

func TestCursorMovement(t *testing.T) {
	const src = "fmt.Println(\"uma linha bem comprida\")\n}\nreturn errors.New(\"outra linha longa\")\n\nabc\n"

	tests := []struct {
		name string
		keys func(ed *Editor)
		want buffer.Pos
	}{
		{
			name: "h na coluna 0 fica parado",
			keys: func(ed *Editor) { press(ed, 'h') },
			want: buffer.Pos{Line: 0, Col: 0},
		},
		{
			name: "k na primeira linha fica parado",
			keys: func(ed *Editor) { press(ed, 'k') },
			want: buffer.Pos{Line: 0, Col: 0},
		},
		{
			name: "l anda uma coluna",
			keys: func(ed *Editor) { press(ed, 'l'); press(ed, 'l') },
			want: buffer.Pos{Line: 0, Col: 2},
		},
		{
			name: "l no ultimo caractere fica parado",
			keys: func(ed *Editor) {
				press(ed, 'j')
				press(ed, 'j')
				press(ed, 'j')
				press(ed, 'j')
				for i := 0; i < 10; i++ {
					press(ed, 'l')
				}
			},
			want: buffer.Pos{Line: 4, Col: 2},
		},
		{
			name: "setas funcionam como hjkl",
			keys: func(ed *Editor) {
				pressCode(ed, input.KeyRight)
				pressCode(ed, input.KeyDown)
				pressCode(ed, input.KeyLeft)
			},
			want: buffer.Pos{Line: 1, Col: 0},
		},
		{
			name: "coluna desejada: longa, curta, longa volta para a coluna original",
			keys: func(ed *Editor) {
				for i := 0; i < 20; i++ {
					press(ed, 'l')
				}
				press(ed, 'j')
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 2, Col: 20},
		},
		{
			name: "coluna desejada: na linha curta o cursor fica no ultimo caractere",
			keys: func(ed *Editor) {
				for i := 0; i < 20; i++ {
					press(ed, 'l')
				}
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 1, Col: 0},
		},
		{
			name: "linha vazia tem coluna 0",
			keys: func(ed *Editor) {
				press(ed, 'l')
				press(ed, 'l')
				press(ed, 'j')
				press(ed, 'j')
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 3, Col: 0},
		},
		{
			name: "h atualiza a coluna desejada",
			keys: func(ed *Editor) {
				for i := 0; i < 20; i++ {
					press(ed, 'l')
				}
				press(ed, 'j')
				press(ed, 'j')
				press(ed, 'h')
				press(ed, 'k')
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 2, Col: 19},
		},
		{
			name: "h que falha na coluna 0 nao muda a coluna desejada (como no vim)",
			keys: func(ed *Editor) {
				for i := 0; i < 20; i++ {
					press(ed, 'l')
				}
				press(ed, 'j')
				press(ed, 'h')
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 2, Col: 20},
		},
		{
			name: "$ vai ao fim e gruda no fim nas proximas linhas",
			keys: func(ed *Editor) {
				press(ed, '$')
				press(ed, 'j')
				press(ed, 'j')
			},
			want: buffer.Pos{Line: 2, Col: 37},
		},
		{
			name: "0 volta ao inicio da linha",
			keys: func(ed *Editor) { press(ed, '$'); press(ed, '0') },
			want: buffer.Pos{Line: 0, Col: 0},
		},
		{
			name: "G vai para a ultima linha e gg volta para a primeira",
			keys: func(ed *Editor) { press(ed, 'G'); press(ed, 'g'); press(ed, 'g') },
			want: buffer.Pos{Line: 0, Col: 0},
		},
		{
			name: "G vai para a ultima linha",
			keys: func(ed *Editor) { press(ed, 'G') },
			want: buffer.Pos{Line: 4, Col: 0},
		},
		{
			name: "PgDn alem do fim para na ultima linha",
			keys: func(ed *Editor) { pressCode(ed, input.KeyPageDown) },
			want: buffer.Pos{Line: 4, Col: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := editorFrom(t, src)
			tt.keys(ed)
			if got := ed.Cursor(); got != tt.want {
				t.Errorf("Cursor() = %+v, esperava %+v", got, tt.want)
			}
		})
	}
}

func TestCursorScreenPosWithTabs(t *testing.T) {
	ed := editorFrom(t, "func main() {\n\t\tx := 1\n}\n")

	press(ed, 'j')
	press(ed, 'l')
	press(ed, 'l')

	frame := ed.Frame()
	if ed.Cursor() != (buffer.Pos{Line: 1, Col: 2}) {
		t.Fatalf("Cursor() = %+v, esperava {Line:1 Col:2}", ed.Cursor())
	}
	gw := ed.gutterWidth()
	if frame.Cursor != (render.ScreenPos{Row: 1, Col: gw + 8}) {
		t.Errorf("Frame.Cursor = %+v, esperava {Row:1 Col:%d} (gutter + duas tabs = 8 celulas)", frame.Cursor, gw+8)
	}
	if got := frameLineText(frame.Lines[1]); []rune(got)[frame.Cursor.Col-gw] != 'x' {
		t.Errorf("na coluna %d do texto esta %q, esperava 'x'", frame.Cursor.Col-gw, string([]rune(got)[frame.Cursor.Col-gw]))
	}
}

func TestStatusShowsCursor(t *testing.T) {
	ed := editorFrom(t, "abc\ndef\n")
	press(ed, 'j')
	press(ed, 'l')

	got := ed.Frame().Status.Text()
	for _, want := range []string{"NORMAL", "teste.go", "2:2", "All"} {
		if !strings.Contains(got, want) {
			t.Errorf("Status = %q, esperava conter %q", got, want)
		}
	}
	if w := len([]rune(got)); w != 80 {
		t.Errorf("status deveria ocupar a largura inteira (80), ocupa %d", w)
	}
}
