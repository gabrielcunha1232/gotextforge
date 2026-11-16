package desktop

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func TestCSSColor(t *testing.T) {
	cases := []struct {
		in   render.Color
		want string
	}{
		{render.Color{}, ""},
		{render.RGB(0x0b, 0x0e, 0x13), "#0b0e13"},
		{render.ANSI(1), "#cd3131"},
		{render.ANSI(16), "#000000"},
		{render.ANSI(196), "#ff0000"},
		{render.ANSI(231), "#ffffff"},
		{render.ANSI(232), "#080808"},
		{render.ANSI(255), "#eeeeee"},
	}
	for _, c := range cases {
		if got := cssColor(c.in); got != c.want {
			t.Errorf("cssColor(%+v) = %q, quero %q", c.in, got, c.want)
		}
	}
}

func TestToFrame(t *testing.T) {
	f := toFrame(render.Frame{
		Lines: []render.StyledLine{{
			{Text: "x := ", Style: render.Style{Fg: render.RGB(1, 2, 3), Bold: true}},
			{Text: ""},
			{Text: "foo", Style: render.Style{Undercurl: true, Sp: render.RGB(0xff, 0, 0), Italic: true}},
		}},
		Status:      render.StyledLine{{Text: "NORMAL", Style: render.Style{Bg: render.RGB(0, 0xad, 0xd8)}}},
		Cursor:      render.ScreenPos{Row: 0, Col: 2},
		CursorShape: render.CursorBar,
	})

	if len(f.Lines) != 1 || len(f.Lines[0]) != 2 {
		t.Fatalf("spans vazios deviam sumir: %+v", f.Lines)
	}
	want := Span{Text: "x := ", Fg: "#010203", Bold: true}
	if f.Lines[0][0] != want {
		t.Errorf("span 0 = %+v, quero %+v", f.Lines[0][0], want)
	}
	want = Span{Text: "foo", Sp: "#ff0000", Italic: true, Undercurl: true}
	if f.Lines[0][1] != want {
		t.Errorf("span 1 = %+v, quero %+v", f.Lines[0][1], want)
	}
	if f.Status[0].Bg != "#00add8" {
		t.Errorf("status bg = %q", f.Status[0].Bg)
	}
	if f.Cursor != (Cursor{Row: 0, Col: 2, Bar: true}) {
		t.Errorf("cursor = %+v", f.Cursor)
	}
}

func TestRunAsync(t *testing.T) {
	events := make(chan any, 3)
	events <- "a"
	events <- "b"
	events <- "c"

	ctx, cancel := context.WithCancel(context.Background())
	var got []string
	done := make(chan struct{})
	go func() {
		runAsync(ctx, events,
			func(ev any) Frame { return Frame{Status: []Span{{Text: ev.(string)}}} },
			func(f Frame) {
				got = append(got, f.Status[0].Text)
				if len(got) == 3 {
					cancel()
				}
			})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runAsync não parou depois do cancel")
	}
	if strings.Join(got, "") != "abc" {
		t.Errorf("emitidos = %v, quero a b c em ordem", got)
	}
}

func TestSeqCresce(t *testing.T) {
	s := NewEditorService(editor.New(nil, ""), nil, nil)
	a := s.Resize(80, 24)
	b := s.Keys([]KeyInput{{Key: "i"}, {Key: "x"}, {Key: "Escape"}})
	if b.Seq <= a.Seq {
		t.Errorf("seq não cresceu: %d depois de %d", b.Seq, a.Seq)
	}
	if len(b.Lines) != 23 {
		t.Errorf("linhas = %d, quero 23 (24 menos a status)", len(b.Lines))
	}
}
