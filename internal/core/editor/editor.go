package editor

import (
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)


type Editor struct {
	lastKey input.Key
	width   int
	height  int
}

func New() *Editor {
	return &Editor{
		width:  80,
		height: 24,
	}
}

func (e *Editor) HandleEvent(ev input.Event) {
	switch ev.Kind {
	case input.EventKey:
		e.lastKey = ev.Key
	case input.EventResize:
		e.width = ev.Width
		e.height = ev.Height
	}
}

func (e *Editor) Frame() render.Frame {
	line := render.StyledLine{
		{Text: "func", Style: render.Style{Fg: render.ColorBlue, Bold: true}},
		{Text: " ", Style: render.Style{}},
		{Text: "main", Style: render.Style{Fg: render.ColorYellow}},
		{Text: "() {", Style: render.Style{}},
	}

	status := "última tecla: (nenhuma)"
	if formatted := e.lastKey.String(); formatted != "" {
		status = "última tecla: " + formatted
	}

	return render.Frame{
		Lines:  []render.StyledLine{line},
		Cursor: render.ScreenPos{Row: 0, Col: 12},
		Status: status,
	}
}
