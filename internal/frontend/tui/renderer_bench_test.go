package tui

import (
	"fmt"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func benchFrame(rows, spansPerLine int) render.Frame {
	styles := []render.Style{
		{},
		{Fg: render.ColorBlue, Bold: true},
		{Fg: render.ColorYellow},
		{Fg: render.RGB(255, 85, 85), Undercurl: true},
	}
	lines := make([]render.StyledLine, rows)
	for i := range lines {
		line := make(render.StyledLine, spansPerLine)
		for j := range line {
			line[j] = render.Span{Text: fmt.Sprintf("tok%d ", j), Style: styles[j%len(styles)]}
		}
		lines[i] = line
	}
	return render.Frame{Lines: lines, Status: render.StyledLine{{Text: " bench.go   1-50/100000"}}}
}

func BenchmarkRenderFrame(b *testing.B) {
	for _, spans := range []int{1, 10, 30} {
		b.Run(fmt.Sprintf("spans_por_linha=%d", spans), func(b *testing.B) {
			frame := benchFrame(50, spans)
			b.ReportAllocs()
			for b.Loop() {
				_ = RenderFrame(frame)
			}
		})
	}
}
