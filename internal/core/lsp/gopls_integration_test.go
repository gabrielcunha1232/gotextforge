package lsp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
)

func waitEventTimeout[T Event](t *testing.T, c *Client, d time.Duration, accept func(T) bool) T {
	t.Helper()
	timeout := time.After(d)
	for {
		select {
		case ev := <-c.Events():
			if st, ok := ev.(StatusEvent); ok && st.Err != nil {
				t.Fatalf("gopls falhou: %v", st.Err)
			}
			if v, ok := ev.(T); ok && (accept == nil || accept(v)) {
				return v
			}
		case <-timeout:
			var zero T
			t.Fatalf("timeout esperando %T do gopls", zero)
			return zero
		}
	}
}

func TestGoplsDeVerdade(t *testing.T) {
	if testing.Short() {
		t.Skip("teste de integracao com gopls (rode sem -short)")
	}
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls nao instalado")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module probe\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "main.go")
	const broken = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"ação\")\n\tvar x int = \"texto\"\n}\n"
	if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	c := Start(path, GoplsStarter)
	defer c.Close()
	c.Open(broken, 1)
	waitEventTimeout[StatusEvent](t, c, 60*time.Second, func(s StatusEvent) bool { return s.Ready })

	diag := waitEventTimeout(t, c, 60*time.Second, func(d DiagnosticsEvent) bool { return len(d.Items) > 0 })
	found := false
	for _, d := range diag.Items {
		t.Logf("diagnostico: linha %d col %d sev %d: %s", d.Range.Start.Line, d.Range.Start.Character, d.Severity, d.Message)
		if d.Range.Start.Line == 6 && d.Severity == SeverityError {
			found = true
		}
	}
	if !found {
		t.Errorf("esperava um erro na linha 7 (indice 6)")
	}

	line := []rune("\tfmt.Println(\"ação\")")
	c.Hover(1, line, buffer.Pos{Line: 5, Col: 6})
	h := waitEventTimeout[HoverEvent](t, c, 30*time.Second, nil)
	t.Logf("hover: %q", h.Text)
	if h.Err != nil || !strings.Contains(h.Text, "Println") {
		t.Errorf("hover = %+v", h)
	}

	const typing = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Pr\n}\n"
	c.Change(typing, 2)
	c.Completion(2, []rune("\tfmt.Pr"), buffer.Pos{Line: 5, Col: 7})
	comp := waitEventTimeout[CompletionEvent](t, c, 30*time.Second, nil)
	labels := make([]string, 0, len(comp.Items))
	for _, it := range comp.Items {
		labels = append(labels, it.Label)
	}
	t.Logf("completion: %v", labels)
	if comp.Err != nil || !strings.Contains(strings.Join(labels, ","), "Println") {
		t.Errorf("completion deveria sugerir Println: %+v", comp)
	}

	const fixed = "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"ok\")\n}\n"
	c.Change(fixed, 3)
	waitEventTimeout(t, c, 60*time.Second, func(d DiagnosticsEvent) bool { return d.Version == 3 && len(d.Items) == 0 })
}
