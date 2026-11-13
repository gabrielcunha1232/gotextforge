package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

func TestParseLocation(t *testing.T) {
	tests := []struct {
		line   string
		file   string
		lineNo int
		col    int
		ok     bool
	}{
		{line: "./main.go:12:5: undefined: x", file: "./main.go", lineNo: 12, col: 5, ok: true},
		{line: `internal\core\editor\ui.go:40:2: syntax error`, file: `internal\core\editor\ui.go`, lineNo: 40, col: 2, ok: true},
		{line: "    editor_test.go:88: texto = \"x\"", file: "editor_test.go", lineNo: 88, col: 1, ok: true},
		{line: `C:\proj\main.go:3:1: erro`, file: `C:\proj\main.go`, lineNo: 3, col: 1, ok: true},
		{line: "docker-compose.yml:7: invalid", file: "docker-compose.yml", lineNo: 7, col: 1, ok: true},
		{line: "ok  	github.com/x/y	0.5s", ok: false},
		{line: "listening on :8080", ok: false},
	}
	for _, tt := range tests {
		file, l, c, ok := parseLocation(tt.line)
		if ok != tt.ok || (ok && (file != tt.file || l != tt.lineNo || c != tt.col)) {
			t.Errorf("parseLocation(%q) = %q %d %d %v", tt.line, file, l, c, ok)
		}
	}
}

func waitTerminal(t *testing.T, ed *Editor) {
	t.Helper()
	deadline := time.After(90 * time.Second)
	for ed.term.job == nil || !ed.term.done {
		select {
		case ev := <-ed.Async():
			ed.HandleAsync(ev)
		case <-deadline:
			t.Fatalf("comando nao terminou; status %q", ed.Frame().Status.Text())
		}
	}
}

func terminalText(ed *Editor) string {
	f := ed.Frame()
	var sb strings.Builder
	for _, l := range f.Lines {
		sb.WriteString(l.Text())
		sb.WriteByte('\n')
	}
	return sb.String()
}

func TestTerminalRodaComandoEMostraSaida(t *testing.T) {
	ed, _ := projectEditor(t, map[string]string{"go.mod": "module x\n\ngo 1.22\n"})
	typeKeys(t, ed, ":!go version<CR>")
	if !ed.term.open || ed.Mode() != ModeNormal {
		t.Fatalf(":! deveria abrir o painel sem tirar o foco do editor (modo %v)", ed.Mode())
	}
	waitTerminal(t, ed)
	out := terminalText(ed)
	for _, want := range []string{"TERMINAL", "$ go version", "go version go", "ok ·"} {
		if !strings.Contains(out, want) {
			t.Errorf("painel deveria conter %q:\n%s", want, out)
		}
	}
	if !strings.Contains(ed.Frame().Status.Text(), "✓ go version") {
		t.Errorf("status = %q", ed.Frame().Status.Text())
	}
}

func TestTerminalErroDeCompilacaoLevaAoArquivo(t *testing.T) {
	ed, root := projectEditor(t, map[string]string{
		"go.mod":       "module x\n\ngo 1.22\n",
		"main.go":      "package main\n\nfunc main() {}\n",
		"util/util.go": "package util\n\nfunc F() int {\n\treturn \"texto\"\n}\n",
	})
	typeKeys(t, ed, " gb")
	waitTerminal(t, ed)
	if ed.term.snap.State != runner.Failed {
		t.Fatalf("go build deveria falhar: %v\n%s", ed.term.snap.State, terminalText(ed))
	}
	if !strings.Contains(ed.Frame().Status.Text(), "saiu com código") {
		t.Errorf("status = %q", ed.Frame().Status.Text())
	}
	if _, _, _, ok := parseLocation(ed.term.snap.Lines[ed.term.sel]); !ok {
		t.Fatalf("ao falhar, a primeira linha de erro fica selecionada: %q", ed.term.snap.Lines)
	}

	typeKeys(t, ed, " tt")
	if ed.Mode() != ModeTerminal || !ed.Frame().HideCursor {
		t.Fatalf("Espaco t t deveria focar o painel")
	}
	typeKeys(t, ed, "<CR>")
	if ed.Mode() != ModeNormal || !samePath(ed.path, filepath.Join(root, "util", "util.go")) {
		t.Fatalf("Enter no erro deveria abrir util/util.go, abriu %q", ed.path)
	}
	if ed.Cursor().Line != 3 {
		t.Errorf("cursor deveria ir para a linha do erro (4), esta em %+v", ed.Cursor())
	}
}

func TestTerminalCancelarEFechar(t *testing.T) {
	ed, _ := projectEditor(t, map[string]string{"x.txt": "x"})
	long := "sleep 30"
	if os.PathSeparator == '\\' {
		long = "ping -n 30 127.0.0.1"
	}
	ed.runCommand(long)
	typeKeys(t, ed, " tt")
	pressCtrl(ed, 'c')
	waitTerminal(t, ed)
	if ed.term.snap.State != runner.Canceled {
		t.Errorf("Ctrl+c no painel deveria cancelar: %v", ed.term.snap.State)
	}
	if !strings.Contains(terminalText(ed), "cancelado") {
		t.Errorf("o painel deveria mostrar cancelado")
	}

	h := ed.viewport.Height
	typeKeys(t, ed, "q")
	if ed.term.open || ed.Mode() != ModeNormal || ed.viewport.Height <= h {
		t.Errorf("q deveria fechar o painel e devolver as linhas ao editor (%d -> %d)", h, ed.viewport.Height)
	}
}

func TestTerminalLayout(t *testing.T) {
	for _, size := range [][2]int{{30, 8}, {80, 24}, {120, 40}} {
		ed, _ := projectEditor(t, map[string]string{"a.go": "package a\n"})
		ed.HandleEvent(input.NewResizeEvent(size[0], size[1]))
		ed.term.last = "go test ./..."
		ed.term.snap = runner.Snapshot{Lines: []string{strings.Repeat("saida longa ", 20), "a.go:1:1: erro"}, State: runner.Failed, ExitCode: 1}
		ed.term.job = &runner.Job{}
		typeKeys(t, ed, " tt")
		f := ed.Frame()
		if len(f.Lines) != size[1]-1 {
			t.Errorf("tela %v: %d linhas, esperava %d", size, len(f.Lines), size[1]-1)
		}
		for i, l := range f.Lines {
			if l.Width() > size[0] {
				t.Errorf("tela %v: linha %d com %d celulas", size, i, l.Width())
			}
		}
		if ed.viewport.Height+ed.terminalRows() != size[1]-1 {
			t.Errorf("tela %v: viewport %d + painel %d deveria ser %d", size, ed.viewport.Height, ed.terminalRows(), size[1]-1)
		}
	}
}

func TestTerminalSemComandoAnterior(t *testing.T) {
	ed, _ := projectEditor(t, map[string]string{})
	typeKeys(t, ed, " tr")
	if !strings.Contains(ed.Frame().Status.Text(), "nenhum comando anterior") {
		t.Errorf("status = %q", ed.Frame().Status.Text())
	}
}

func TestTerminalAvisaInterativo(t *testing.T) {
	ed, _ := projectEditor(t, map[string]string{})
	ed.start = func(string, string, func(bool)) (*runner.Job, error) { return &runner.Job{}, nil }
	ed.runCommand("docker exec -it app sh")
	if !strings.Contains(ed.Frame().Status.Text(), "interativos") {
		t.Errorf("deveria avisar que programas interativos precisam do terminal: %q", ed.Frame().Status.Text())
	}
}
