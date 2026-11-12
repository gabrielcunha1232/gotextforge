package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

var theme = [...]render.Style{
	syntax.KindPlain:       {},
	syntax.KindKeyword:     {Fg: render.ColorMagenta},
	syntax.KindString:      {Fg: render.ColorGreen},
	syntax.KindNumber:      {Fg: render.ColorYellow},
	syntax.KindComment:     {Fg: render.ColorBrightBlack, Italic: true},
	syntax.KindPredeclared: {Fg: render.ColorCyan},
	syntax.KindFunc:        {Fg: render.ColorBlue},
}

func (e *Editor) renderBufferLine(i int) render.StyledLine {
	runes := e.buf.Line(i)
	if !e.highlight {
		return view.RenderLine(runes, plainStyle, e.width, view.DefaultTabStop)
	}

	tokens, _ := syntax.HighlightLine(runes, e.syntax.StateAt(e.buf, i))
	styles := make([]render.Style, len(runes))
	for _, t := range tokens {
		for j := t.Start; j < t.End; j++ {
			styles[j] = theme[t.Kind]
		}
	}
	return view.RenderLine(runes, func(j int) render.Style { return styles[j] }, e.width, view.DefaultTabStop)
}

func plainStyle(int) render.Style {
	return render.Style{}
}
