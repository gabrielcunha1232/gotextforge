package editor

import (
	"strings"
	"testing"
)

func TestMolduraExternaSoTexto(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.SetExternalChrome(true)

	f := ed.Frame()
	if len(f.Lines) != 23 {
		t.Fatalf("linhas = %d, quero 23 (24 menos a linha de comando)", len(f.Lines))
	}
	if !strings.Contains(f.Lines[0].Text(), "package b") {
		t.Errorf("a primeira linha deveria ser o texto, sem a linha de abas: %q", f.Lines[0].Text())
	}
	if f.Cursor.Row != 0 {
		t.Errorf("cursor na linha %d, quero 0", f.Cursor.Row)
	}
	if !strings.Contains(f.Status.Text(), `"b.go"`) {
		t.Errorf("a mensagem de arquivo aberto deveria aparecer na última linha: %q", f.Status.Text())
	}
	typeKeys(t, ed, "l")
	if s := ed.Frame().Status.Text(); s != "" {
		t.Errorf("sem comando nem mensagem, a última linha deveria ficar vazia: %q", s)
	}
	if ed.viewport.Height != 23 {
		t.Errorf("viewport = %d, quero 23", ed.viewport.Height)
	}
}

func TestMolduraExternaLinhaDeComando(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n"}, "a.go")
	ed.SetExternalChrome(true)

	typeKeys(t, ed, ":wq")
	f := ed.Frame()
	if f.Status.Text() != ":wq" {
		t.Errorf("linha de comando = %q, quero %q", f.Status.Text(), ":wq")
	}
	if f.Cursor.Row != len(f.Lines) || f.Cursor.Col != 3 {
		t.Errorf("cursor = %+v, quero linha %d coluna 3", f.Cursor, len(f.Lines))
	}

	typeKeys(t, ed, "<Esc>:w<CR>")
	if msg := ed.Frame().Status.Text(); !strings.Contains(msg, "a.go") {
		t.Errorf("depois do :w a mensagem deveria aparecer na última linha: %q", msg)
	}
	typeKeys(t, ed, "j")
	if msg := ed.Frame().Status.Text(); msg != "" {
		t.Errorf("a mensagem deveria sumir na tecla seguinte: %q", msg)
	}
}

func TestMolduraExternaIgnoraPaineis(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n"}, "a.go")
	ed.SetRoot(t.TempDir())
	ed.SetExternalChrome(true)
	ed.ShowExplorer()
	ed.term.open = true
	ed.layout()

	f := ed.Frame()
	if len(f.Lines) != 23 {
		t.Errorf("terminal aberto não deveria tirar linhas do texto: %d", len(f.Lines))
	}
	if !strings.HasPrefix(strings.TrimLeft(f.Lines[0].Text(), " "), "1") {
		t.Errorf("explorador aberto não deveria aparecer à esquerda: %q", f.Lines[0].Text())
	}
	if f.Cursor.Col != ed.gutterWidth() {
		t.Errorf("cursor.Col = %d, quero %d (sem deslocamento do explorador)", f.Cursor.Col, ed.gutterWidth())
	}
}

func TestMolduraInternaContinuaIgual(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	f := ed.Frame()
	if strings.Contains(f.Lines[0].Text(), "package") {
		t.Errorf("sem a moldura externa, a primeira linha continua sendo a das abas: %q", f.Lines[0].Text())
	}
	if !strings.Contains(f.Status.Text(), "NORMAL") {
		t.Errorf("sem a moldura externa, a status line continua: %q", f.Status.Text())
	}
}
