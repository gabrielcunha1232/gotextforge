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
	var b strings.Builder
	for _, span := range line {
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

	t.Run("tecla j desce 1 linha", func(t *testing.T) {
		buf := makeNumberedBuffer(t, 100)
		ed := New(buf, "teste.go")
		ed.HandleEvent(input.NewResizeEvent(80, 11))

		ed.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyRune, Rune: 'j'}))

		frame := ed.Frame()
		if got := frameLineText(frame.Lines[0]); got != "linha 2" {
			t.Errorf("primeira linha apos 'j' = %q, esperava %q", got, "linha 2")
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
