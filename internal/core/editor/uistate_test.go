package editor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

func TestUIAbasEStatus(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.txt": "um\r\ndois\r\n"}, "a.go", "b.txt")

	typeKeys(t, ed, "jA!<Esc>")
	ui := ed.UI()
	if len(ui.Tabs) != 2 || ui.Tabs[0].Name != "a.go" || ui.Tabs[0].Active || !ui.Tabs[1].Active || !ui.Tabs[1].Modified || ui.Tabs[0].Modified {
		t.Errorf("abas = %+v", ui.Tabs)
	}
	s := ui.Status
	if s.File != "b.txt" || s.Line != 2 || s.Col != 5 || !s.Modified || s.LineEnding != "CRLF" || s.Language != "txt" {
		t.Errorf("status = %+v", s)
	}
	if ui.Mode != "NORMAL" || ui.Focus != "editor" {
		t.Errorf("modo/foco = %q/%q", ui.Mode, ui.Focus)
	}

	typeKeys(t, ed, "d")
	if p := ed.UI().Status.Pending; p != "d" {
		t.Errorf("pendente = %q, quero %q", p, "d")
	}
	typeKeys(t, ed, "<Esc>i")
	if ui := ed.UI(); ui.Mode != "INSERT" || ui.Focus != "editor" {
		t.Errorf("modo/foco = %q/%q", ui.Mode, ui.Focus)
	}
	typeKeys(t, ed, "<Esc>:")
	if f := ed.UI().Focus; f != "command" {
		t.Errorf("foco no : = %q", f)
	}
}

func TestUIProblemasConverteUTF16(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\nvar s = \"😀\" + x\n"}, "a.go")
	ed.HandleAsync(lsp.DiagnosticsEvent{
		URI: lsp.PathToURI("a.go"),
		Items: []lsp.Diagnostic{
			{Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 15}}, Severity: lsp.SeverityError, Message: "undefined: x"},
			{Range: lsp.Range{Start: lsp.Position{Line: 0, Character: 0}}, Severity: lsp.SeverityWarning, Message: "aviso"},
		},
	})

	p := ed.UI().Problems
	if len(p) != 2 {
		t.Fatalf("problemas = %+v", p)
	}
	if p[0].Line != 2 || p[0].Col != 15 || p[0].Severity != "error" || p[0].Name != "a.go" || p[0].Message != "undefined: x" {
		t.Errorf("o emoji ocupa 2 unidades UTF-16 e 1 rune: quero linha 2 coluna 15, veio %+v", p[0])
	}
	if p[1].Severity != "warning" {
		t.Errorf("severidade = %q", p[1].Severity)
	}
	if s := ed.UI().Status; s.Errors != 1 || s.Warnings != 1 {
		t.Errorf("contadores = E%d W%d", s.Errors, s.Warnings)
	}
}

func TestUIExploradorEVersaoDoGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.26.1\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644)
	os.Mkdir(filepath.Join(dir, "vazia"), 0o755)

	ed := New(nil, "")
	ed.HandleEvent(input.NewResizeEvent(100, 30))
	ed.SetRoot(dir)
	ed.ShowExplorer()
	ed.OpenFile(filepath.Join(dir, "main.go"))

	x := ed.UI().Explorer
	if !x.Open || x.Name != filepath.Base(dir) {
		t.Errorf("explorador = open %v nome %q", x.Open, x.Name)
	}
	var names []string
	for _, r := range x.Rows {
		names = append(names, r.Name)
		if r.Name == "main.go" && !r.Current {
			t.Errorf("main.go está aberto e deveria vir marcado como atual")
		}
		if r.Name == "vazia" && !r.Dir {
			t.Errorf("vazia deveria ser pasta")
		}
	}
	if got := strings.Join(names, ","); got != "vazia,go.mod,main.go" {
		t.Errorf("linhas = %s", got)
	}
	if v := ed.UI().Status.GoVersion; v != "1.26" {
		t.Errorf("versão do Go = %q, quero 1.26", v)
	}

	typeKeys(t, ed, "  ")
	typeKeys(t, ed, "mai")
	f := ed.UI().Finder
	if f == nil || f.Query != "mai" || len(f.Items) != 1 || f.Items[0].Text != "main.go" || f.Total != 2 {
		t.Fatalf("busca = %+v", f)
	}
	if fmt.Sprint(f.Items[0].Positions) != "[0 1 2]" {
		t.Errorf("posições = %v", f.Items[0].Positions)
	}
	typeKeys(t, ed, "x")
	if f := ed.UI().Finder; f == nil || !f.Create || len(f.Items) != 0 {
		t.Errorf("sem resultado, a busca deveria oferecer criar: %+v", f)
	}
	typeKeys(t, ed, "<Esc>")
	if ed.UI().Finder != nil {
		t.Errorf("busca fechada deveria virar nil")
	}
}

func TestUITerminalEDocker(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n"}, "a.go")

	lines := make([]string, 2500)
	for i := range lines {
		lines[i] = fmt.Sprintf("linha %d", i)
	}
	ed.term.open = true
	ed.term.job = &runner.Job{Command: "go test ./...", Dir: "/proj"}
	ed.term.snap = runner.Snapshot{Lines: lines, State: runner.Failed, ExitCode: 1}

	ui := ed.UI()
	tm := ui.Terminal
	if ui.Bottom != "terminal" || tm.Command != "go test ./..." || tm.State != "failed" || tm.ExitCode != 1 {
		t.Errorf("terminal = %+v", tm)
	}
	if tm.First != 500 || len(tm.Lines) != 2000 || tm.Lines[0] != "linha 500" {
		t.Errorf("deveria mandar só as últimas 2000 linhas: first %d, %d linhas, primeira %q", tm.First, len(tm.Lines), tm.Lines[0])
	}

	ed.dock.open = true
	ed.dock.tab = dockImages
	ed.dock.images = []docker.Image{{Repository: "golang", Tag: "1.27"}}
	ed.dock.err = errors.New("Docker Desktop não está rodando")
	ui = ed.UI()
	if ui.Bottom != "docker" || ui.Docker.Tab != "Imagens" || len(ui.Docker.Images) != 1 || ui.Docker.Error == "" {
		t.Errorf("docker = %+v (bottom %q)", ui.Docker, ui.Bottom)
	}
}

func BenchmarkUI(b *testing.B) {
	var src strings.Builder
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&src, "linha %d\n", i)
	}
	ed := New(nil, "")
	ed.load = fakeFiles(map[string]string{"a.go": src.String(), "b.go": "package b\n"})
	ed.HandleEvent(input.NewResizeEvent(200, 50))
	ed.OpenFile("a.go")
	ed.OpenFile("b.go")
	lines := make([]string, 5000)
	for i := range lines {
		lines[i] = fmt.Sprintf("arquivo.go:%d:5: alguma mensagem de erro do compilador", i)
	}
	ed.term.open = true
	ed.term.job = &runner.Job{Command: "go build ./..."}
	ed.term.snap = runner.Snapshot{Lines: lines}

	var size int
	for b.Loop() {
		data, err := json.Marshal(ed.UI())
		if err != nil {
			b.Fatal(err)
		}
		size = len(data)
	}
	b.ReportMetric(float64(size), "B/ui")
}
