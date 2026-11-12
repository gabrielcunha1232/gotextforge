package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func BenchmarkQuadroCompleto(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&sb, "\tif err := process(ctx, %d, \"item\"); err != nil { return fmt.Errorf(\"falha: %%w\", err) } // nota\n", i)
	}
	buf, err := buffer.Load(strings.NewReader(sb.String()))
	if err != nil {
		b.Fatal(err)
	}
	ed := editor.New(buf, "bench.go")
	ed.HandleEvent(input.NewResizeEvent(200, 51))
	down := input.NewKeyEvent(input.Key{Rune: 'j'})

	b.ReportAllocs()
	for b.Loop() {
		ed.HandleEvent(down)
		_ = RenderFrame(ed.Frame())
	}
}
