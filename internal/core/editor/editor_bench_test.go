package editor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func syntheticGoSource(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		switch i % 4 {
		case 0:
			fmt.Fprintf(&b, "func handler%d(w http.ResponseWriter, r *http.Request) {\n", i)
		case 1:
			fmt.Fprintf(&b, "\tif err := process(r.Context(), %d); err != nil {\n", i)
		case 2:
			fmt.Fprintf(&b, "\t\treturn fmt.Errorf(\"falha ao processar item %%d: %%w\", %d, err) // comentário com ação\n", i)
		default:
			b.WriteString("\t}\n")
		}
	}
	return b.String()
}

func benchEditor(b *testing.B, lines int) *Editor {
	b.Helper()
	buf, err := buffer.Load(strings.NewReader(syntheticGoSource(lines)))
	if err != nil {
		b.Fatal(err)
	}
	ed := New(buf, "bench.go")
	ed.HandleEvent(input.NewResizeEvent(200, 51))
	return ed
}

func BenchmarkFrame(b *testing.B) {
	for _, lines := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("linhas=%d", lines), func(b *testing.B) {
			ed := benchEditor(b, lines)
			ed.HandleEvent(input.NewKeyEvent(input.Key{Rune: 'G'}))
			b.ReportAllocs()
			for b.Loop() {
				_ = ed.Frame()
			}
		})
	}
}

func BenchmarkScrollAndFrame(b *testing.B) {
	ed := benchEditor(b, 100_000)
	down := input.NewKeyEvent(input.Key{Rune: 'j'})
	b.ReportAllocs()
	for b.Loop() {
		ed.HandleEvent(down)
		_ = ed.Frame()
	}
}
