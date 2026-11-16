package desktop

import (
	"fmt"
	"sync"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type Span struct {
	Text      string `json:"t"`
	Fg        string `json:"fg,omitempty"`
	Bg        string `json:"bg,omitempty"`
	Sp        string `json:"sp,omitempty"`
	Bold      bool   `json:"b,omitempty"`
	Italic    bool   `json:"i,omitempty"`
	Underline bool   `json:"u,omitempty"`
	Undercurl bool   `json:"c,omitempty"`
}

type Cursor struct {
	Row  int  `json:"row"`
	Col  int  `json:"col"`
	Bar  bool `json:"bar,omitempty"`
	Hide bool `json:"hide,omitempty"`
}

type Frame struct {
	Lines  [][]Span `json:"lines"`
	Status []Span   `json:"status"`
	Cursor Cursor   `json:"cursor"`
}

type EditorService struct {
	mu   sync.Mutex
	ed   *editor.Editor
	quit func()
}

func NewEditorService(ed *editor.Editor, quit func()) *EditorService {
	return &EditorService{ed: ed, quit: quit}
}

func (s *EditorService) Resize(cols, rows int) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ed.HandleEvent(input.NewResizeEvent(cols, rows))
	return toFrame(s.ed.Frame())
}

func (s *EditorService) Keys(keys []KeyInput) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range keys {
		key, ok := translateKey(k)
		if !ok {
			continue
		}
		s.ed.HandleEvent(input.NewKeyEvent(key))
		if s.ed.ShouldQuit() {
			if s.quit != nil {
				go s.quit()
			}
			break
		}
	}
	return toFrame(s.ed.Frame())
}

func toFrame(f render.Frame) Frame {
	lines := make([][]Span, len(f.Lines))
	for i, l := range f.Lines {
		lines[i] = toSpans(l)
	}
	return Frame{
		Lines:  lines,
		Status: toSpans(f.Status),
		Cursor: Cursor{
			Row:  f.Cursor.Row,
			Col:  f.Cursor.Col,
			Bar:  f.CursorShape == render.CursorBar,
			Hide: f.HideCursor,
		},
	}
}

func toSpans(l render.StyledLine) []Span {
	spans := make([]Span, 0, len(l))
	for _, s := range l {
		if s.Text == "" {
			continue
		}
		spans = append(spans, Span{
			Text:      s.Text,
			Fg:        cssColor(s.Style.Fg),
			Bg:        cssColor(s.Style.Bg),
			Sp:        cssColor(s.Style.Sp),
			Bold:      s.Style.Bold,
			Italic:    s.Style.Italic,
			Underline: s.Style.Underline,
			Undercurl: s.Style.Undercurl,
		})
	}
	return spans
}

func cssColor(c render.Color) string {
	switch c.Kind {
	case render.ColorRGB:
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	case render.ColorANSI:
		r, g, b := ansiRGB(c.Index)
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return ""
}

var ansiBase = [16][3]uint8{
	{0x00, 0x00, 0x00}, {0xcd, 0x31, 0x31}, {0x0d, 0xbc, 0x79}, {0xe5, 0xe5, 0x10},
	{0x24, 0x72, 0xc8}, {0xbc, 0x3f, 0xbc}, {0x11, 0xa8, 0xcd}, {0xe5, 0xe5, 0xe5},
	{0x66, 0x66, 0x66}, {0xf1, 0x4c, 0x4c}, {0x23, 0xd1, 0x8b}, {0xf5, 0xf5, 0x43},
	{0x3b, 0x8e, 0xea}, {0xd6, 0x70, 0xd6}, {0x29, 0xb8, 0xdb}, {0xff, 0xff, 0xff},
}

func ansiRGB(i uint8) (uint8, uint8, uint8) {
	switch {
	case i < 16:
		c := ansiBase[i]
		return c[0], c[1], c[2]
	case i < 232:
		n := i - 16
		level := func(v uint8) uint8 {
			if v == 0 {
				return 0
			}
			return 55 + v*40
		}
		return level(n / 36), level(n / 6 % 6), level(n % 6)
	default:
		v := 8 + (i-232)*10
		return v, v, v
	}
}
