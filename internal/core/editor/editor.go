package editor

import (
	"fmt"
	"path/filepath"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/view"
)

type Editor struct {
	buf      *buffer.Buffer
	viewport view.Viewport
	filename string
	width    int
	height   int
}

func New(buf *buffer.Buffer, filename string) *Editor {
	if buf == nil {
		buf = buffer.New()
	}
	if filename == "" {
		filename = "[Sem Nome]"
	} else {
		filename = filepath.Base(filename)
	}

	const defaultWidth = 80
	const defaultHeight = 24

	return &Editor{
		buf:      buf,
		viewport: view.NewViewport(defaultHeight - 1),
		filename: filename,
		width:    defaultWidth,
		height:   defaultHeight,
	}
}

func (e *Editor) HandleEvent(ev input.Event) {
	switch ev.Kind {
	case input.EventResize:
		e.width = ev.Width
		e.height = ev.Height
		vpHeight := e.height - 1
		if vpHeight < 1 {
			vpHeight = 1
		}
		e.viewport.Resize(vpHeight, e.buf.LineCount())

	case input.EventKey:
		e.handleKey(ev.Key)
	}
}

func (e *Editor) handleKey(k input.Key) {
	total := e.buf.LineCount()

	switch {
	case k.Code == input.KeyDown || (k.Code == input.KeyRune && k.Rune == 'j' && k.Mod == 0):
		e.viewport.Scroll(1, total)

	case k.Code == input.KeyUp || (k.Code == input.KeyRune && k.Rune == 'k' && k.Mod == 0):
		e.viewport.Scroll(-1, total)

	case k.Code == input.KeyPageDown:
		e.viewport.Scroll(e.viewport.Height, total)

	case k.Code == input.KeyPageUp:
		e.viewport.Scroll(-e.viewport.Height, total)

	case k.Code == input.KeyHome || (k.Code == input.KeyRune && k.Rune == 'g' && k.Mod == 0):
		e.viewport.ScrollTop()

	case k.Code == input.KeyEnd || (k.Code == input.KeyRune && k.Rune == 'G' && k.Mod == 0):
		e.viewport.ScrollBottom(total)
	}
}

func (e *Editor) Frame() render.Frame {
	total := e.buf.LineCount()
	start, end := e.viewport.VisibleRange(total)

	vpHeight := e.viewport.Height
	lines := make([]render.StyledLine, 0, vpHeight)

	for i := start; i < end; i++ {
		formatted := view.FormatLine(e.buf.Line(i), e.width, view.DefaultTabStop)
		lines = append(lines, render.StyledLine{
			{Text: formatted, Style: render.Style{}},
		})
	}

	tildeStyle := render.Style{Fg: render.ColorBrightBlack}
	for len(lines) < vpHeight {
		lines = append(lines, render.StyledLine{
			{Text: "~", Style: tildeStyle},
		})
	}

	firstVisible := start + 1
	lastVisible := end
	if total == 0 {
		firstVisible = 0
	}
	status := fmt.Sprintf(" %s   %d-%d/%d", e.filename, firstVisible, lastVisible, total)

	return render.Frame{
		Lines:  lines,
		Cursor: render.ScreenPos{Row: 0, Col: 0},
		Status: status,
	}
}
