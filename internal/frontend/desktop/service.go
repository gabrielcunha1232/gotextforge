package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/wailsapp/wails/v3/pkg/application"
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

type Pos struct {
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	Progress string `json:"progress"`
	Pending  string `json:"pending,omitempty"`
}

type Frame struct {
	Seq    uint64          `json:"seq"`
	Lines  [][]Span        `json:"lines"`
	Status []Span          `json:"status"`
	Cursor Cursor          `json:"cursor"`
	Pos    Pos             `json:"pos"`
	UI     json.RawMessage `json:"ui,omitempty"`
}

const FrameEvent = "frame"

type EditorService struct {
	mu     sync.Mutex
	ed     *editor.Editor
	seq    uint64
	quit   func()
	emit   func(Frame)
	cancel context.CancelFunc
	lastUI []byte

	pickDir func(start string) (string, error)
}

func NewEditorService(ed *editor.Editor, quit func(), emit func(Frame)) *EditorService {
	return &EditorService{ed: ed, quit: quit, emit: emit}
}

func (s *EditorService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	ctx, s.cancel = context.WithCancel(ctx)
	go runAsync(ctx, s.ed.Async(), s.handleAsync, s.emit)
	return nil
}

func (s *EditorService) ServiceShutdown() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *EditorService) Start(cols, rows int) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastUI = nil
	s.ed.HandleEvent(input.NewResizeEvent(cols, rows))
	return s.frame()
}

func (s *EditorService) Theme() map[string]string {
	out := map[string]string{}
	for name, c := range editor.Palette() {
		out[name] = cssColor(c)
	}
	return out
}

func (s *EditorService) Resize(cols, rows int) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ed.HandleEvent(input.NewResizeEvent(cols, rows))
	return s.frame()
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
		if s.checkQuit() {
			break
		}
	}
	return s.frame()
}

func (s *EditorService) SwitchTab(i int) Frame {
	return s.act(func() { s.ed.SwitchTab(i) })
}

func (s *EditorService) CloseTab(i int) Frame {
	return s.act(func() { s.ed.CloseTab(i) })
}

func (s *EditorService) ExplorerClick(path string) Frame {
	return s.act(func() { s.ed.ExplorerClick(path) })
}

func (s *EditorService) TogglePanel(name string) Frame {
	return s.act(func() { s.ed.TogglePanel(name) })
}

func (s *EditorService) GoTo(path string, line, col int) Frame {
	return s.act(func() { s.ed.GoTo(path, line, col) })
}

func (s *EditorService) TerminalClick(i int) Frame {
	return s.act(func() { s.ed.TerminalClick(i) })
}

func (s *EditorService) TerminalAction(name string) Frame {
	return s.act(func() { s.ed.TerminalAction(name) })
}

func (s *EditorService) DockerTab(name string) Frame {
	return s.act(func() { s.ed.DockerTab(name) })
}

func (s *EditorService) DockerClick(i int) Frame {
	return s.act(func() { s.ed.DockerClick(i) })
}

func (s *EditorService) DockerKey(key string) Frame {
	return s.act(func() { s.ed.DockerKey(key) })
}

func (s *EditorService) FinderClick(i int) Frame {
	return s.act(func() { s.ed.FinderClick(i) })
}

func (s *EditorService) ClickText(row, col int) Frame {
	return s.act(func() { s.ed.ClickText(row, col) })
}

func (s *EditorService) ScrollText(delta int) Frame {
	return s.act(func() { s.ed.ScrollText(delta) })
}

func (s *EditorService) OpenFolder() Frame {
	if s.pickDir == nil {
		return s.act(func() {})
	}
	s.mu.Lock()
	start := s.ed.Root()
	s.mu.Unlock()

	dir, err := s.pickDir(start)
	if err != nil || dir == "" {
		return s.act(func() {})
	}
	return s.act(func() { s.ed.OpenFolder(dir) })
}

func (s *EditorService) allowClose() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ed.ShouldQuit() {
		return true
	}
	s.ed.RequestQuit()
	if s.ed.ShouldQuit() {
		return true
	}
	if s.emit != nil {
		s.emit(s.frame())
	}
	return false
}

func (s *EditorService) act(fn func()) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
	s.checkQuit()
	return s.frame()
}

func (s *EditorService) handleAsync(ev any) Frame {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ed.HandleAsync(ev)
	s.checkQuit()
	return s.frame()
}

func (s *EditorService) checkQuit() bool {
	if !s.ed.ShouldQuit() {
		return false
	}
	if s.quit != nil {
		go s.quit()
	}
	return true
}

func (s *EditorService) frame() Frame {
	s.seq++
	f := toFrame(s.ed.Frame())
	f.Seq = s.seq

	ui := s.ed.UI()
	f.Pos = Pos{Line: ui.Status.Line, Col: ui.Status.Col, Progress: ui.Status.Progress, Pending: ui.Status.Pending}
	ui.Status.Line, ui.Status.Col, ui.Status.Progress, ui.Status.Pending = 0, 0, "", ""
	if data, err := json.Marshal(ui); err == nil && !bytes.Equal(data, s.lastUI) {
		s.lastUI = data
		f.UI = data
	}
	return f
}

func runAsync(ctx context.Context, events <-chan any, handle func(any) Frame, emit func(Frame)) {
	if events == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			f := handle(ev)
			if emit != nil {
				emit(f)
			}
		}
	}
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

func (s *EditorService) BenchMode() bool {
	return os.Getenv("GTF_BENCH") != ""
}

func (s *EditorService) Report(text string) {
	fmt.Fprintln(os.Stderr, text)
	if s.quit != nil {
		go s.quit()
	}
}
