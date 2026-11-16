package desktop

import (
	"context"
	"os"
	"path/filepath"
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

func TestUISoQuandoMuda(t *testing.T) {
	s := NewEditorService(editor.New(nil, ""), nil, nil)
	f := s.Start(80, 24)
	if len(f.UI) == 0 {
		t.Fatal("o Start deveria mandar o retrato inteiro")
	}
	f = s.Keys([]KeyInput{{Key: "i"}, {Key: "a"}, {Key: "Escape"}})
	if len(f.UI) == 0 {
		t.Error("o modo e o [+] mudaram; o retrato deveria vir")
	}
	f = s.Keys([]KeyInput{{Key: "h"}})
	if len(f.UI) != 0 {
		t.Errorf("só o cursor mexeu; o retrato não deveria vir: %s", f.UI)
	}
	if f.Pos.Col != 1 || f.Pos.Line != 1 {
		t.Errorf("pos = %+v, quero linha 1 coluna 1", f.Pos)
	}
	if f = s.Start(80, 24); len(f.UI) == 0 {
		t.Error("um novo Start (página recarregada) deveria mandar o retrato de novo")
	}
}

func TestTheme(t *testing.T) {
	th := NewEditorService(editor.New(nil, ""), nil, nil).Theme()
	if th["bg"] != "#0b0e13" || th["mode-normal"] == "" || th["mode-insert"] == "" {
		t.Errorf("tema = %v", th)
	}
}

func TestAcoesDeAba(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")
	os.WriteFile(a, []byte("package a\n"), 0o644)
	os.WriteFile(b, []byte("package b\n"), 0o644)

	ed := editor.New(nil, "")
	ed.OpenFile(a)
	ed.OpenFile(b)
	s := NewEditorService(ed, nil, nil)
	s.Start(80, 24)

	f := s.SwitchTab(0)
	if !strings.Contains(string(f.UI), `"Name":"a.go","Path":`) || ed.UI().Status.File != "a.go" {
		t.Fatalf("SwitchTab(0) deveria ativar a.go: %s", f.UI)
	}
	f = s.CloseTab(0)
	if tabs := ed.UI().Tabs; len(tabs) != 1 || tabs[0].Name != "b.go" || len(f.UI) == 0 {
		t.Errorf("CloseTab(0) deveria deixar só b.go: %+v", tabs)
	}
}

func TestAllowClose(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.go")
	os.WriteFile(a, []byte("package a\n"), 0o644)
	ed := editor.New(nil, "")
	ed.OpenFile(a)
	var emitted int
	s := NewEditorService(ed, nil, func(Frame) { emitted++ })
	s.Start(80, 24)

	if !s.allowClose() {
		t.Fatal("sem alterações, o X deveria fechar")
	}

	ed = editor.New(nil, "")
	ed.OpenFile(a)
	s = NewEditorService(ed, nil, func(Frame) { emitted++ })
	s.Start(80, 24)
	s.Keys([]KeyInput{{Key: "A"}, {Key: "x"}, {Key: "Escape"}})
	if s.allowClose() || emitted != 1 {
		t.Fatalf("com alteração, o X deveria ser cancelado e a pergunta enviada: emitidos %d", emitted)
	}
	s.Keys([]KeyInput{{Key: "n"}, {Key: "Enter"}})
	if !ed.ShouldQuit() || !s.allowClose() {
		t.Error("respondendo não, o editor sai e o X é liberado")
	}
}

func TestOpenFolder(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	ed := editor.New(nil, "")
	ed.SetRoot(a)
	s := NewEditorService(ed, nil, nil)
	s.Start(80, 24)

	var start string
	answer := ""
	s.pickDir = func(dir string) (string, error) {
		start = dir
		if !s.mu.TryLock() {
			t.Error("o diálogo não pode rodar com o mutex travado")
		} else {
			s.mu.Unlock()
		}
		return answer, nil
	}

	s.OpenFolder()
	if start != ed.Root() || ed.Root() != a {
		t.Fatalf("cancelar não deveria mudar nada: começou em %q, raiz %q", start, ed.Root())
	}

	answer = b
	f := s.OpenFolder()
	if ed.Root() != b {
		t.Fatalf("raiz %q, quero %q", ed.Root(), b)
	}
	if !strings.Contains(string(f.UI), `"Root":`) {
		t.Error("trocar de pasta muda o retrato; ele deveria vir no quadro")
	}
}
