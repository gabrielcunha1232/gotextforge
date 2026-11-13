package editor

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/engine"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

type terminalTick struct{}

type StartFunc func(command, dir string, notify func(final bool)) (*runner.Job, error)

type terminal struct {
	open   bool
	job    *runner.Job
	snap   runner.Snapshot
	sel    int
	top    int
	follow bool
	last   string
	dir    string
	done   bool
	cwd    string
	prev   string
}

var locationRe = regexp.MustCompile(`((?:[A-Za-z]:)?[^\s:"'()\[\]]+\.[A-Za-z0-9]+):(\d+)(?::(\d+))?`)

func parseLocation(line string) (file string, lineNo, col int, ok bool) {
	m := locationRe.FindStringSubmatch(line)
	if m == nil {
		return "", 0, 0, false
	}
	lineNo, _ = strconv.Atoi(m[2])
	col = 1
	if m[3] != "" {
		col, _ = strconv.Atoi(m[3])
	}
	return m[1], lineNo, col, lineNo > 0
}

func (e *Editor) bottomPanelHeight() int {
	rows := min(max(e.height/3, 6), 15)
	rows = min(rows, e.height-3)
	if rows < 2 {
		return 0
	}
	return rows
}

// bottomRows is the height of the single bottom panel slot, shared by the
// terminal and the docker panel (they are mutually exclusive).
func (e *Editor) bottomRows() int {
	if e.dock.open || e.term.open {
		return e.bottomPanelHeight()
	}
	return 0
}

func (e *Editor) terminalRows() int {
	if !e.term.open || e.dock.open {
		return 0
	}
	return e.bottomPanelHeight()
}

func (e *Editor) openTerminal(focus bool) {
	if e.dock.open {
		e.closeDocker()
	}
	if !e.term.open {
		e.term.open = true
		e.layout()
	}
	if focus {
		e.mode = ModeTerminal
	}
}

func (e *Editor) closeTerminal() {
	e.term.open = false
	if e.mode == ModeTerminal {
		e.mode = ModeNormal
	}
	e.layout()
}

func (e *Editor) toggleTerminal() {
	switch {
	case !e.term.open:
		e.openTerminal(true)
	case e.mode == ModeTerminal:
		e.closeTerminal()
	default:
		e.mode = ModeTerminal
	}
}

func (e *Editor) promptCommand() {
	label := shortenPath(sanitize(e.terminalDir()), 40) + " $ "
	e.openPrompt(label, "", e.runCommand)
}

func (e *Editor) terminalDir() string {
	if e.term.cwd != "" {
		return e.term.cwd
	}
	return e.projectRoot()
}

func (e *Editor) terminalBuiltin(command string) bool {
	if strings.ContainsAny(command, "&|;<>") {
		return false
	}
	fields := strings.Fields(command)
	switch fields[0] {
	case "pwd":
		e.openTerminal(false)
		e.message = "pasta do terminal: " + e.terminalDir()
		return true
	case "cd", "chdir":
	default:
		return false
	}

	args := fields[1:]
	if len(args) > 0 && strings.EqualFold(args[0], "/d") {
		args = args[1:]
	}
	target := strings.Trim(strings.Join(args, " "), `"'`)
	current := e.terminalDir()

	switch {
	case target == "":
		target = e.projectRoot()
	case target == "-":
		if e.term.prev == "" {
			e.message = "cd -: não há pasta anterior"
			return true
		}
		target = e.term.prev
	case target == "~" || strings.HasPrefix(target, "~/") || strings.HasPrefix(target, `~\`):
		home, err := os.UserHomeDir()
		if err != nil {
			e.message = "cd: não achei a pasta do usuário"
			return true
		}
		target = filepath.Join(home, target[1:])
	case !filepath.IsAbs(target):
		target = filepath.Join(current, target)
	}
	target = filepath.Clean(target)

	if fi, err := os.Stat(target); err != nil || !fi.IsDir() {
		e.message = "cd: pasta não encontrada: " + strings.Join(args, " ")
		e.openTerminal(false)
		return true
	}

	e.term.prev = current
	e.term.cwd = target
	if samePath(target, e.projectRoot()) {
		e.term.cwd = ""
	}
	e.openTerminal(false)
	e.message = "pasta do terminal: " + target
	return true
}

func (e *Editor) rerunCommand() {
	if e.term.last == "" {
		e.message = "nenhum comando anterior (use Espaço t n ou :!comando)"
		return
	}
	e.runCommand(e.term.last)
}

func (e *Editor) notifyTerminal(final bool) {
	if final {
		go func() {
			select {
			case e.async <- terminalTick{}:
			case <-time.After(5 * time.Second):
			}
		}()
		return
	}
	select {
	case e.async <- terminalTick{}:
	default:
	}
}

func (e *Editor) runCommand(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}
	if e.terminalBuiltin(command) {
		return
	}
	if e.term.job != nil {
		e.term.job.Cancel()
	}
	dir := e.terminalDir()
	job, err := e.start(command, dir, e.notifyTerminal)
	if err != nil {
		e.message = "erro ao executar: " + err.Error()
		return
	}
	e.term = terminal{open: e.term.open, job: job, follow: true, last: command, dir: dir, cwd: e.term.cwd, prev: e.term.prev}
	e.openTerminal(false)

	padded := " " + command + " "
	switch {
	case strings.Contains(padded, " -it ") || strings.Contains(padded, " -ti "):
		e.message = "este painel não recebe teclado: programas interativos (como docker exec -it) precisam do seu terminal"
	case e.firstModified() != nil:
		e.message = "atenção: há arquivos não salvos, e o comando usa a versão do disco (:wa salva tudo)"
	default:
		e.message = "$ " + command
	}
}

func (e *Editor) receiveTerminal(terminalTick) {
	t := &e.term
	if t.job == nil {
		return
	}
	t.snap = t.job.Snapshot()
	n := len(t.snap.Lines)
	if t.follow {
		t.sel = max(n-1, 0)
	}
	if t.snap.State == runner.Running || t.done {
		return
	}
	t.done = true
	e.refreshTree()
	if e.dock.open && e.dock.pending {
		e.dock.pending = false
		e.startDockerRefresh()
	}
	elapsed := t.snap.Elapsed.Round(100 * time.Millisecond)
	switch t.snap.State {
	case runner.Succeeded:
		e.message = fmt.Sprintf("✓ %s (%v)", t.last, elapsed)
	case runner.Canceled:
		e.message = "comando cancelado: " + t.last
	default:
		for i, line := range t.snap.Lines {
			if _, _, _, ok := parseLocation(line); ok {
				t.sel, t.follow = i, false
				break
			}
		}
		e.message = fmt.Sprintf("✗ %s saiu com código %d (Enter no erro abre o arquivo)", t.last, t.snap.ExitCode)
	}
}

func (e *Editor) feedLeader(k input.Key) bool {
	if !(k.Code == input.KeyRune && k.Rune == ' ' && k.Mod == 0) && e.parser.Pending() == "" {
		return false
	}
	if cmd, done := e.parser.Feed(k); done && cmd.Action != engine.ActionNone {
		e.runAction(cmd)
	}
	return true
}

func (e *Editor) resolveLocation(file string) (string, bool) {
	candidates := []string{file}
	if !filepath.IsAbs(file) {
		candidates = []string{filepath.Join(e.term.dir, file), filepath.Join(e.projectRoot(), file)}
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c, true
		}
	}
	suffix := string(filepath.Separator) + filepath.Clean(filepath.FromSlash(file))
	found := ""
	_ = filepath.WalkDir(e.projectRoot(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, suffix) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found, found != ""
}

func (e *Editor) jumpToLocation(i int) {
	if i < 0 || i >= len(e.term.snap.Lines) {
		return
	}
	file, line, col, ok := parseLocation(e.term.snap.Lines[i])
	if !ok {
		return
	}
	path, ok := e.resolveLocation(file)
	if !ok {
		e.message = fmt.Sprintf("arquivo %q não encontrado no projeto", file)
		return
	}
	e.OpenFile(path)
	target := e.buf.ClampPos(buffer.Pos{Line: line - 1})
	e.cursor.Line = target.Line
	e.setCol(min(max(col-1, 0), e.maxCol(target.Line)))
	e.mode = ModeNormal
}

func (e *Editor) nextLocation(from, step int) int {
	lines := e.term.snap.Lines
	for i := from + step; i >= 0 && i < len(lines); i += step {
		if _, _, _, ok := parseLocation(lines[i]); ok {
			return i
		}
	}
	return from
}

func (e *Editor) handleTerminalKey(k input.Key) {
	if e.feedLeader(k) {
		return
	}
	t := &e.term
	n := len(t.snap.Lines)
	last := max(n-1, 0)
	page := max(e.terminalRows()-2, 1)
	plain := func(r rune) bool { return k.Code == input.KeyRune && k.Rune == r && k.Mod == 0 }

	switch {
	case plain('j') || k.Code == input.KeyDown:
		t.sel = min(t.sel+1, last)
		t.follow = t.sel == last
	case plain('k') || k.Code == input.KeyUp:
		t.sel = max(t.sel-1, 0)
		t.follow = false
	case plain('g') || k.Code == input.KeyHome:
		t.sel, t.follow = 0, false
	case plain('G') || k.Code == input.KeyEnd:
		t.sel, t.follow = last, true
	case k.Code == input.KeyPageDown:
		t.sel = min(t.sel+page, last)
		t.follow = t.sel == last
	case k.Code == input.KeyPageUp:
		t.sel, t.follow = max(t.sel-page, 0), false
	case plain('n'):
		t.sel, t.follow = e.nextLocation(t.sel, 1), false
	case plain('N'):
		t.sel, t.follow = e.nextLocation(t.sel, -1), false
	case k.Code == input.KeyEnter || plain('o'):
		e.jumpToLocation(t.sel)
	case plain('r'):
		e.rerunCommand()
	case plain('i') || plain(':'):
		e.promptCommand()
	case isCtrl(k, 'c'):
		if t.job != nil && t.snap.State == runner.Running {
			t.job.Cancel()
		}
	case plain('q'):
		e.closeTerminal()
	case k.Code == input.KeyEsc:
		e.mode = ModeNormal
	}
}

func terminalLineStyle(line string) render.Style {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "--- FAIL"), strings.HasPrefix(trimmed, "FAIL"), strings.HasPrefix(trimmed, "panic:"),
		strings.HasPrefix(trimmed, "Error"), strings.HasPrefix(trimmed, "error"):
		return render.Style{Fg: colRed}
	case strings.HasPrefix(trimmed, "--- PASS"), strings.HasPrefix(trimmed, "ok "), trimmed == "PASS", strings.HasPrefix(trimmed, "✔"):
		return render.Style{Fg: colGreen}
	case strings.Contains(strings.ToLower(trimmed), "warning"):
		return render.Style{Fg: colYellow}
	}
	if _, _, _, ok := parseLocation(line); ok {
		return render.Style{Fg: colRed}
	}
	return render.Style{Fg: colFgDark}
}

func (e *Editor) terminalPanel(rows, width int) []render.StyledLine {
	t := &e.term
	bar := render.Style{Fg: colFgDark, Bg: colBgStatus}

	var color render.Color
	var status string
	elapsed := t.snap.Elapsed.Round(100 * time.Millisecond)
	switch {
	case t.job == nil:
		color, status = colFgGutter, "Espaço t n ou :!comando para executar"
	case t.snap.State == runner.Running:
		color, status = colYellow, fmt.Sprintf("rodando %v · Ctrl+c cancela", elapsed)
	case t.snap.State == runner.Succeeded:
		color, status = colGreen, fmt.Sprintf("ok · %v", elapsed)
	case t.snap.State == runner.Canceled:
		color, status = colDark3, "cancelado"
	default:
		color, status = colRed, fmt.Sprintf("saiu com código %d · %v", t.snap.ExitCode, elapsed)
	}

	header := render.StyledLine{{Text: " TERMINAL ", Style: render.Style{Fg: colBlack, Bg: color, Bold: true}}}
	dir := e.terminalDir()
	header = append(header, render.Span{Text: " " + shortenPath(sanitize(dir), max(width/3, 16)) + " ", Style: render.Style{Fg: colBlue, Bg: colBgStatus}})
	if t.last != "" {
		header = append(header, render.Span{Text: "$ " + sanitize(t.last) + " ", Style: render.Style{Fg: colFg, Bg: colBgStatus, Bold: true}})
	}
	right := render.Span{Text: " " + status + " ", Style: bar}
	if gap := width - header.Width() - (render.StyledLine{right}).Width(); gap > 0 {
		header = append(header, render.Span{Text: strings.Repeat(" ", gap), Style: bar}, right)
	}
	out := []render.StyledLine{padTo(truncateLine(header, width), width, bar)}

	body := rows - 1
	lines := t.snap.Lines
	if t.sel < t.top {
		t.top = t.sel
	}
	if t.sel >= t.top+body {
		t.top = t.sel - body + 1
	}
	t.top = max(min(t.top, len(lines)-body), 0)

	focused := e.mode == ModeTerminal
	for i := t.top; i < t.top+body; i++ {
		var line render.StyledLine
		switch {
		case i < len(lines):
			style := terminalLineStyle(lines[i])
			if focused && i == t.sel {
				style.Bg = colBgSel
			}
			text := " " + sanitize(lines[i])
			line = padTo(truncateLine(render.StyledLine{{Text: text, Style: style}}, width), width, render.Style{Bg: style.Bg})
		case i == 0 && t.job != nil && t.snap.State == runner.Running:
			line = render.StyledLine{{Text: " aguardando saída…", Style: render.Style{Fg: colDark3, Italic: true}}}
		case i == 0 && t.job != nil:
			line = render.StyledLine{{Text: " (o comando não escreveu nada)", Style: render.Style{Fg: colDark3, Italic: true}}}
		}
		out = append(out, truncateLine(line, width))
	}
	return out
}

func shortenPath(path string, limit int) string {
	r := []rune(path)
	if len(r) <= limit || limit < 2 {
		return path
	}
	return "…" + string(r[len(r)-limit+1:])
}
