package editor

import (
	"path/filepath"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

const uiTerminalLines = 2000

type UIState struct {
	Mode     string
	Theme    string
	Focus    string
	Root     string
	Tabs     []UITab
	Explorer UIExplorer
	Bottom   string
	Terminal UITerminal
	Docker   UIDocker
	Problems []UIProblem
	Finder   *UIFinder
	Status   UIStatus
}

type UITab struct {
	Name     string
	Path     string
	Modified bool
	Active   bool
}

type UIExplorer struct {
	Open     bool
	Name     string
	Rows     []UITreeRow
	Selected int
}

type UITreeRow struct {
	Name     string
	Path     string
	Depth    int
	Dir      bool
	Expanded bool
	Empty    bool
	Current  bool
}

type UITerminal struct {
	Open     bool
	Command  string
	Dir      string
	State    string
	ExitCode int
	Millis   int64
	First    int
	Lines    []string
	Kinds    string
	Selected int
	Follow   bool
}

type UIDocker struct {
	Open       bool
	Tab        string
	Selected   int
	Containers []docker.Container
	Services   []docker.ComposeService
	Images     []docker.Image
	Error      string
}

type UIProblem struct {
	Path     string
	Name     string
	Line     int
	Col      int
	Severity string
	Message  string
}

type UIFinder struct {
	Title     string
	Query     string
	Items     []UIFinderItem
	Create    bool
	Selected  int
	Total     int
	Truncated bool
}

type UIFinderItem struct {
	Text      string
	Positions []int
}

type UIStatus struct {
	File       string
	Path       string
	Modified   bool
	Line       int
	Col        int
	Progress   string
	Pending    string
	Errors     int
	Warnings   int
	Language   string
	LSP        string
	LineEnding string
	BOM        bool
	GoVersion  string
}

func (e *Editor) UI() UIState {
	return UIState{
		Mode:     modeNames[e.mode],
		Theme:    e.theme.Name,
		Focus:    e.focusName(),
		Root:     e.projectRoot(),
		Tabs:     e.uiTabs(),
		Explorer: e.uiExplorer(),
		Bottom:   e.uiBottom(),
		Terminal: e.uiTerminal(),
		Docker:   e.uiDocker(),
		Problems: e.uiProblems(),
		Finder:   e.uiFinder(),
		Status:   e.uiStatus(),
	}
}

func (e *Editor) focusName() string {
	switch e.mode {
	case ModeExplorer:
		return "explorer"
	case ModeTerminal:
		return "terminal"
	case ModeDocker:
		return "docker"
	case ModeFinder:
		return "finder"
	case ModeCommand, ModeSearch, ModePrompt:
		return "command"
	}
	return "editor"
}

func (e *Editor) uiTabs() []UITab {
	tabs := make([]UITab, len(e.docs))
	for i, d := range e.docs {
		tabs[i] = UITab{
			Name:     d.filename,
			Path:     d.path,
			Modified: d.buf.Modified(),
			Active:   d == e.document,
		}
	}
	return tabs
}

func (e *Editor) uiExplorer() UIExplorer {
	x := &e.explorer
	if x.tree == nil {
		return UIExplorer{Open: x.open}
	}
	rows := make([]UITreeRow, len(x.rows))
	for i, n := range x.rows {
		rows[i] = UITreeRow{
			Name:     n.name,
			Path:     n.path,
			Depth:    n.depth,
			Dir:      n.dir,
			Expanded: n.expanded,
			Empty:    n.dir && n.loaded && len(n.children) == 0,
			Current:  !n.dir && e.path != "" && samePath(n.path, e.path),
		}
	}
	return UIExplorer{Open: x.open, Name: x.tree.name, Rows: rows, Selected: x.sel}
}

func (e *Editor) uiBottom() string {
	switch {
	case e.dock.open:
		return "docker"
	case e.term.open:
		return "terminal"
	}
	return ""
}

func (e *Editor) uiTerminal() UITerminal {
	t := &e.term
	out := UITerminal{Open: t.open, Dir: e.terminalDir(), Selected: t.sel, Follow: t.follow}
	if t.job == nil {
		return out
	}
	out.Command = t.job.Command
	out.Dir = t.job.Dir
	out.ExitCode = t.snap.ExitCode
	out.Millis = t.snap.Elapsed.Milliseconds()
	switch t.snap.State {
	case runner.Running:
		out.State = "running"
	case runner.Succeeded:
		out.State = "ok"
	case runner.Failed:
		out.State = "failed"
	case runner.Canceled:
		out.State = "canceled"
	}
	lines := t.snap.Lines
	kinds := e.terminalKinds()
	if len(lines) > uiTerminalLines {
		out.First = len(lines) - uiTerminalLines
		lines = lines[out.First:]
		kinds = kinds[out.First:]
	}
	out.Lines = append([]string(nil), lines...)
	out.Kinds = string(kinds)
	return out
}

func (e *Editor) uiDocker() UIDocker {
	d := &e.dock
	out := UIDocker{
		Open:       d.open,
		Tab:        dockerTabNames[d.tab],
		Selected:   d.sel,
		Containers: append([]docker.Container(nil), d.containers...),
		Services:   append([]docker.ComposeService(nil), d.services...),
		Images:     append([]docker.Image(nil), d.images...),
	}
	if d.err != nil {
		out.Error = d.err.Error()
	}
	return out
}

func severityName(s lsp.Severity) string {
	switch s {
	case lsp.SeverityError:
		return "error"
	case lsp.SeverityWarning:
		return "warning"
	case lsp.SeverityInfo:
		return "info"
	}
	return "hint"
}

func (e *Editor) uiProblems() []UIProblem {
	var out []UIProblem
	for _, d := range e.docs {
		for _, diag := range d.diags {
			line := diag.Range.Start.Line
			col := 0
			if line < d.buf.LineCount() {
				col = lsp.FromPosition(d.buf.Line(line), diag.Range.Start).Col
			}
			out = append(out, UIProblem{
				Path:     d.path,
				Name:     d.filename,
				Line:     line + 1,
				Col:      col + 1,
				Severity: severityName(diag.Severity),
				Message:  diag.Message,
			})
		}
	}
	return out
}

func (e *Editor) uiFinder() *UIFinder {
	f := e.finder
	if e.mode != ModeFinder || f == nil {
		return nil
	}
	items := make([]UIFinderItem, len(f.matches))
	for i, m := range f.matches {
		items[i] = UIFinderItem{Text: m.Text, Positions: append([]int(nil), m.Positions...)}
	}
	return &UIFinder{
		Title:     strings.TrimSpace(f.title),
		Query:     string(f.query),
		Items:     items,
		Create:    f.createRow(),
		Selected:  f.sel,
		Total:     len(f.items),
		Truncated: f.truncated,
	}
}

func (e *Editor) cachedGoVersion() string {
	root := e.projectRoot()
	if e.goVersion[0] != root || e.goVersion[1] == "" {
		e.goVersion = [2]string{root, goModVersion(root)}
	}
	return e.goVersion[1]
}

func (e *Editor) uiStatus() UIStatus {
	errs, warns := e.diagnosticCounts()
	s := UIStatus{
		File:       e.filename,
		Path:       e.path,
		Modified:   e.buf.Modified(),
		Line:       e.cursor.Line + 1,
		Col:        e.cursor.Col + 1,
		Progress:   e.progress(),
		Pending:    e.parser.Pending(),
		Errors:     errs,
		Warnings:   warns,
		LineEnding: "LF",
		BOM:        e.buf.HasBOM(),
		GoVersion:  e.cachedGoVersion(),
	}
	if e.buf.LineEnding() == buffer.LineEndingCRLF {
		s.LineEnding = "CRLF"
	}
	if e.highlight {
		s.Language = "go"
	} else if ext := filepath.Ext(e.path); ext != "" {
		s.Language = strings.TrimPrefix(ext, ".")
	}
	switch {
	case e.lsp.client == nil:
		s.LSP = ""
	case e.lsp.failed:
		s.LSP = "failed"
	case !e.lsp.ready:
		s.LSP = "starting"
	default:
		s.LSP = "ready"
	}
	return s
}
