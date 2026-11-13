package editor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type dockerTab int

const (
	dockContainers dockerTab = iota
	dockCompose
	dockImages
)

var dockerTabNames = []string{"Containers", "Compose", "Imagens"}

type dockerState struct {
	open       bool
	tab        dockerTab
	sel        int
	top        int
	containers []docker.Container
	images     []docker.Image
	services   []docker.ComposeService
	err        error
	updated    time.Time
	stop       chan struct{}
	gen        int
}

type dockerSnapshot struct {
	gen        int
	containers []docker.Container
	images     []docker.Image
	services   []docker.ComposeService
	err        error
}

func defaultDockerExec(dir string) docker.Exec {
	return func(args ...string) ([]byte, error) {
		cmd := exec.Command("docker", args...)
		cmd.Dir = dir
		return cmd.Output()
	}
}

func (e *Editor) toggleDocker() {
	switch {
	case !e.dock.open:
		e.openDocker()
	case e.mode == ModeDocker:
		e.closeDocker()
	default:
		e.mode = ModeDocker
	}
}

func (e *Editor) openDocker() {
	if e.dockerExec == nil {
		e.message = "docker desativado"
		return
	}
	e.dock.open = true
	e.dock.tab = dockContainers
	e.dock.sel = 0
	e.dock.top = 0
	e.mode = ModeDocker
	e.startDockerRefresh()
	e.layout()
}

func (e *Editor) closeDocker() {
	e.dock.open = false
	e.stopDockerRefresh()
	if e.mode == ModeDocker {
		e.mode = ModeNormal
	}
	e.layout()
}

func (e *Editor) startDockerRefresh() {
	e.stopDockerRefresh()
	if e.dockerExec == nil {
		return
	}
	client := docker.New(e.dockerExec(e.projectRoot()))
	stop := make(chan struct{})
	e.dock.stop = stop
	e.dock.gen++
	go dockerRefreshLoop(client, e.dock.gen, stop, e.async)
}

func (e *Editor) stopDockerRefresh() {
	if e.dock.stop != nil {
		close(e.dock.stop)
		e.dock.stop = nil
	}
}

func dockerRefreshLoop(c *docker.Client, gen int, stop <-chan struct{}, async chan<- any) {
	fetch := func() {
		snap := dockerSnapshot{gen: gen}
		if _, err := c.Info(); err != nil {
			snap.err = err
		} else {
			snap.containers, _ = c.ListContainers()
			snap.images, _ = c.ListImages()
			snap.services, _ = c.ComposePS()
		}
		select {
		case async <- snap:
		case <-stop:
		}
	}
	fetch()
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			fetch()
		}
	}
}

func (e *Editor) receiveDocker(s dockerSnapshot) {
	if s.gen != e.dock.gen {
		return
	}
	e.dock.containers = s.containers
	e.dock.images = s.images
	e.dock.services = s.services
	e.dock.err = s.err
	e.dock.updated = time.Now()
	if e.dock.sel >= e.dockerItemCount() {
		e.dock.sel = max(e.dockerItemCount()-1, 0)
	}
}

func (e *Editor) dockerItemCount() int {
	switch e.dock.tab {
	case dockCompose:
		return len(e.dock.services)
	case dockImages:
		return len(e.dock.images)
	default:
		return len(e.dock.containers)
	}
}

func (e *Editor) dockerRows() int {
	if !e.dock.open {
		return 0
	}
	return e.bottomPanelHeight()
}

func (e *Editor) handleDockerKey(k input.Key) {
	if e.feedLeader(k) {
		return
	}
	d := &e.dock
	last := max(e.dockerItemCount()-1, 0)
	plain := func(r rune) bool { return k.Code == input.KeyRune && k.Rune == r && k.Mod == 0 }

	switch {
	case plain('j') || k.Code == input.KeyDown:
		d.sel = min(d.sel+1, last)
	case plain('k') || k.Code == input.KeyUp:
		d.sel = max(d.sel-1, 0)
	case plain('g') || k.Code == input.KeyHome:
		d.sel = 0
	case plain('G') || k.Code == input.KeyEnd:
		d.sel = last
	case (k.Code == input.KeyTab && k.Mod == 0) || k.Code == input.KeyRight:
		d.tab = (d.tab + 1) % 3
		d.sel, d.top = 0, 0
	case (k.Code == input.KeyTab && k.Mod == input.ModShift) || k.Code == input.KeyLeft:
		d.tab = (d.tab + 2) % 3
		d.sel, d.top = 0, 0
	case plain('q'):
		e.closeDocker()
	case k.Code == input.KeyEsc:
		e.mode = ModeNormal
	}
}

func dockerStateColor(state string) render.Color {
	switch {
	case strings.HasPrefix(state, "running"):
		return colGreen
	case strings.HasPrefix(state, "paused"), strings.HasPrefix(state, "created"):
		return colYellow
	default:
		return colRed
	}
}

func dockerDot(state string) string {
	if strings.HasPrefix(state, "running") {
		return "●"
	}
	return "○"
}

func (e *Editor) dockerPanel(rows, width int) []render.StyledLine {
	d := &e.dock
	bar := render.Style{Fg: colFgDark, Bg: colBgStatus}

	header := render.StyledLine{{Text: " DOCKER ", Style: render.Style{Fg: colBlack, Bg: colBlue1, Bold: true}}}
	for i, name := range dockerTabNames {
		style := render.Style{Fg: colDark3, Bg: colBgStatus}
		if dockerTab(i) == d.tab {
			style = render.Style{Fg: colFg, Bg: colBgSel, Bold: true}
		}
		header = append(header, render.Span{Text: " " + name + " ", Style: style})
	}
	right := " Tab troca · q fecha "
	if !d.updated.IsZero() {
		right = fmt.Sprintf(" atualizado há %ds · Tab troca · q fecha ", int(time.Since(d.updated).Seconds()))
	}
	rightSpan := render.Span{Text: right, Style: bar}
	if gap := width - header.Width() - (render.StyledLine{rightSpan}).Width(); gap > 0 {
		header = append(header, render.Span{Text: strings.Repeat(" ", gap), Style: bar}, rightSpan)
	}
	out := []render.StyledLine{padTo(truncateLine(header, width), width, bar)}

	body := rows - 1
	lines := e.dockerBody()

	if d.err != nil {
		msg := render.StyledLine{{Text: " " + sanitize(d.err.Error()), Style: render.Style{Fg: colRed}}}
		out = append(out, truncateLine(msg, width))
		for len(out) < rows {
			out = append(out, render.StyledLine{})
		}
		return out
	}

	if d.sel < d.top {
		d.top = d.sel
	}
	if d.sel >= d.top+body {
		d.top = d.sel - body + 1
	}
	d.top = max(min(d.top, len(lines)-body), 0)

	for i := d.top; i < d.top+body; i++ {
		var line render.StyledLine
		switch {
		case i < len(lines):
			line = lines[i]
			if i == d.sel {
				line = withBgSpans(line, colBgSel)
			}
			line = padTo(truncateLine(line, width), width, render.Style{Bg: selBg(i == d.sel)})
		case i == 0:
			line = render.StyledLine{{Text: " (nada aqui)", Style: render.Style{Fg: colDark3, Italic: true}}}
		}
		out = append(out, truncateLine(line, width))
	}
	return out
}

func selBg(sel bool) render.Color {
	if sel {
		return colBgSel
	}
	return render.Color{}
}

func withBgSpans(line render.StyledLine, bg render.Color) render.StyledLine {
	out := make(render.StyledLine, len(line))
	for i, s := range line {
		s.Style.Bg = bg
		out[i] = s
	}
	return out
}

func (e *Editor) dockerBody() []render.StyledLine {
	d := &e.dock
	var out []render.StyledLine
	switch d.tab {
	case dockCompose:
		for _, s := range d.services {
			state := s.State
			if s.State == "exited" {
				state = fmt.Sprintf("exited (%d)", s.ExitCode)
			}
			out = append(out, render.StyledLine{
				{Text: " " + dockerDot(s.State) + " ", Style: render.Style{Fg: dockerStateColor(s.State)}},
				{Text: pad(s.Service, 14), Style: render.Style{Fg: colFg}},
				{Text: pad(state, 18), Style: render.Style{Fg: dockerStateColor(s.State)}},
				{Text: s.Ports, Style: render.Style{Fg: colFgDark}},
			})
		}
	case dockImages:
		for _, img := range d.images {
			repo := img.Repository
			if img.Tag != "" {
				repo += ":" + img.Tag
			}
			out = append(out, render.StyledLine{
				{Text: "   " + pad(repo, 34), Style: render.Style{Fg: colFg}},
				{Text: pad(img.Size, 12), Style: render.Style{Fg: colFgDark}},
				{Text: shortID(img.ID), Style: render.Style{Fg: colDark3}},
			})
		}
	default:
		for _, c := range d.containers {
			out = append(out, render.StyledLine{
				{Text: " " + dockerDot(c.State) + " ", Style: render.Style{Fg: dockerStateColor(c.State)}},
				{Text: pad(c.Name, 18), Style: render.Style{Fg: colFg}},
				{Text: pad(sanitize(c.Status), 26), Style: render.Style{Fg: dockerStateColor(c.State)}},
				{Text: c.Ports, Style: render.Style{Fg: colFgDark}},
			})
		}
	}
	return out
}

func pad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		if w < 2 {
			return string(r[:w])
		}
		return string(r[:w-1]) + " "
	}
	return s + strings.Repeat(" ", w-len(r))
}

func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
