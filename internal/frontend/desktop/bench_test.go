package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
)

func bigService(b *testing.B, lines int) *EditorService {
	var sb strings.Builder
	sb.WriteString("package main\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\n")
	for i := 0; strings.Count(sb.String(), "\n") < lines; i++ {
		fmt.Fprintf(&sb, "// processa%d junta os nomes /* não é bloco */\n", i)
		fmt.Fprintf(&sb, "func processa%d(nomes []string, limite int) (string, int) {\n", i)
		sb.WriteString("\ttotal := 0\n\tvar b strings.Builder\n\tfor i, n := range nomes {\n")
		fmt.Fprintf(&sb, "\t\tb.WriteString(fmt.Sprintf(\"%%d:%%s;\", i+%d, n))\n", i)
		sb.WriteString("\t\ttotal += len(n) * 0x1f\n\t}\n\treturn b.String(), total\n}\n\n")
	}
	path := filepath.Join(b.TempDir(), "big.go")
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		b.Fatal(err)
	}
	ed := editor.New(nil, "")
	ed.OpenFile(path)
	s := NewEditorService(ed, nil, nil)
	s.Resize(200, 50)
	return s
}

func reportJSON(b *testing.B, f Frame) {
	data, err := json.Marshal(f)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(len(data)), "B/frame")
}

func BenchmarkKeysScroll5k(b *testing.B) {
	s := bigService(b, 5000)
	var f Frame
	i := 0
	for b.Loop() {
		key := "j"
		if i%4000 == 3999 {
			s.Keys([]KeyInput{{Key: "g"}, {Key: "g"}})
		}
		f = s.Keys([]KeyInput{{Key: key}})
		if _, err := json.Marshal(f); err != nil {
			b.Fatal(err)
		}
		i++
	}
	reportJSON(b, f)
}

func BenchmarkKeysTyping5k(b *testing.B) {
	s := bigService(b, 5000)
	s.Keys([]KeyInput{{Key: "2"}, {Key: "5"}, {Key: "0"}, {Key: "0"}, {Key: "G"}, {Key: "A"}})
	var f Frame
	i := 0
	for b.Loop() {
		key := "x"
		if i%2 == 1 {
			key = "Backspace"
		}
		f = s.Keys([]KeyInput{{Key: key}})
		if _, err := json.Marshal(f); err != nil {
			b.Fatal(err)
		}
		i++
	}
	reportJSON(b, f)
}
