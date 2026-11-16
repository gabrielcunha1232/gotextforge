package desktop

import (
	"sync"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type Cursor struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Frame struct {
	Lines  []string `json:"lines"`
	Status string   `json:"status"`
	Cursor Cursor   `json:"cursor"`
}

type EditorService struct {
	mu sync.Mutex
	ed *editor.Editor
}

func NewEditorService(ed *editor.Editor) *EditorService {
	return &EditorService{ed: ed}
}

func (s *EditorService) Resize(cols, rows int) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ed.HandleEvent(input.NewResizeEvent(cols, rows))
	return toFrame(s.ed.Frame())
}

func toFrame(f render.Frame) Frame {
	lines := make([]string, len(f.Lines))
	for i, l := range f.Lines {
		lines[i] = l.Text()
	}
	return Frame{
		Lines:  lines,
		Status: f.Status.Text(),
		Cursor: Cursor{Row: f.Cursor.Row, Col: f.Cursor.Col},
	}
}
