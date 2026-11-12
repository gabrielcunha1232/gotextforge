package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
)

func pumpUntil(t *testing.T, ed *Editor, what string, cond func() bool) {
	t.Helper()
	deadline := time.After(60 * time.Second)
	for !cond() {
		select {
		case ev := <-ed.Async():
			ed.HandleAsync(ev)
		case <-deadline:
			t.Fatalf("timeout esperando: %s (status %q)", what, ed.Frame().Status.Text())
		}
	}
}

func frameDump(ed *Editor) string {
	var sb strings.Builder
	f := ed.Frame()
	for _, l := range f.Lines {
		sb.WriteString(strings.TrimRight(l.Text(), " "))
		sb.WriteByte('\n')
	}
	sb.WriteString(f.Status.Text())
	return sb.String()
}

func TestEditorComGoplsDeVerdade(t *testing.T) {
	if testing.Short() {
		t.Skip("integracao com gopls (rode sem -short)")
	}
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls nao instalado")
	}

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module probe\n\ngo 1.22\n"), 0o644)
	path := filepath.Join(dir, "main.go")
	const src = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"ok\")\n}\n"
	os.WriteFile(path, []byte(src), 0o644)

	buf, err := buffer.Load(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	ed := New(buf, path)
	ed.HandleEvent(input.NewResizeEvent(90, 20))
	ed.SetLanguageClient(lsp.Start(path, lsp.GoplsStarter))
	defer ed.Close()

	pumpUntil(t, ed, "gopls pronto", func() bool { return ed.lsp.ready })

	typeKeys(t, ed, "jjjjjo\tvar x int = \"texto\"<Esc>")
	pumpUntil(t, ed, "erro aparecer", func() bool {
		errs, _ := ed.diagnosticCounts()
		return errs > 0
	})
	f := ed.Frame()
	under := false
	for _, s := range f.Lines[6] {
		if s.Style.Undercurl && s.Style.Sp == colRed {
			under = true
		}
	}
	if !under {
		t.Errorf("a linha com erro deveria ter sublinhado vermelho:\n%s", frameDump(ed))
	}
	if !strings.Contains(f.Lines[6].Text(), "●") {
		t.Errorf("a linha com erro deveria ter a mensagem no fim:\n%s", frameDump(ed))
	}
	t.Logf("com erro:\n%s", frameDump(ed))

	typeKeys(t, ed, "u")
	pumpUntil(t, ed, "erro sumir depois do undo", func() bool {
		errs, _ := ed.diagnosticCounts()
		return errs == 0
	})

	typeKeys(t, ed, "gg5jwwwK")
	pumpUntil(t, ed, "hover", func() bool { return ed.hover != nil })
	dump := frameDump(ed)
	t.Logf("hover:\n%s", dump)
	if !strings.Contains(dump, "func fmt.Println") || !strings.Contains(dump, "╭") {
		t.Errorf("janela de hover deveria mostrar a assinatura de Println")
	}
	cursorRow := ed.Frame().Cursor.Row
	if !strings.Contains(ed.Frame().Lines[cursorRow].Text(), "fmt.Println(\"ok\")") {
		t.Errorf("a janela de hover nao pode cobrir a linha do cursor")
	}
	typeKeys(t, ed, "<Esc>")

	typeKeys(t, ed, "o\tfmt.Pr")
	pumpUntil(t, ed, "menu de completion", func() bool { return ed.comp != nil })
	labels := []string{}
	for _, it := range ed.comp.items {
		labels = append(labels, it.Label)
	}
	t.Logf("completion: %v\n%s", labels, frameDump(ed))
	if !strings.Contains(strings.Join(labels, ","), "Println") {
		t.Errorf("menu deveria sugerir Println: %v", labels)
	}
}
