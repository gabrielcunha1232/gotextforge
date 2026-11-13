package editor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

const sampleDockerPS = `{"ID":"abc123","Image":"postgres:16","Names":"db","Ports":"0.0.0.0:5432->5432/tcp","State":"running","Status":"Up 2 hours"}
{"ID":"def456","Image":"myapp/worker","Names":"worker","Ports":"","State":"exited","Status":"Exited (1) 3 minutes ago"}
`

func fakeDockerExec(info error) func(string) docker.Exec {
	return func(string) docker.Exec {
		return func(args ...string) ([]byte, error) {
			switch args[0] {
			case "info":
				if info != nil {
					return nil, info
				}
				return []byte("29.9.1\n"), nil
			case "ps":
				return []byte(sampleDockerPS), nil
			case "images":
				return []byte(""), nil
			case "compose":
				return []byte("[]"), nil
			}
			return nil, nil
		}
	}
}

func drainOne(t *testing.T, ed *Editor) {
	t.Helper()
	select {
	case ev := <-ed.Async():
		ed.HandleAsync(ev)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout esperando o snapshot do docker")
	}
}

func frameText(ed *Editor) string {
	var b strings.Builder
	for _, line := range ed.Frame().Lines {
		b.WriteString(frameLineText(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func TestDockerPanelShowsContainers(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.dockerExec = fakeDockerExec(nil)
	ed.HandleEvent(input.NewResizeEvent(80, 30))

	ed.openDocker()
	drainOne(t, ed)

	if len(ed.dock.containers) != 2 {
		t.Fatalf("esperava 2 containers, veio %d", len(ed.dock.containers))
	}
	text := frameText(ed)
	if !strings.Contains(text, "DOCKER") || !strings.Contains(text, "worker") {
		t.Errorf("painel não mostrou os containers:\n%s", text)
	}
	if !strings.Contains(text, "Exited (1)") {
		t.Errorf("painel não mostrou o status do container que saiu com erro:\n%s", text)
	}
}

func TestDockerPanelDaemonDown(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.dockerExec = fakeDockerExec(errors.New("no daemon"))
	ed.HandleEvent(input.NewResizeEvent(80, 30))

	ed.openDocker()
	drainOne(t, ed)

	if !errors.Is(ed.dock.err, docker.ErrDaemonDown) {
		t.Fatalf("esperava ErrDaemonDown no estado, veio %v", ed.dock.err)
	}
	if len(ed.dock.containers) != 0 {
		t.Errorf("com o daemon desligado não deveria listar containers")
	}
	if !strings.Contains(frameText(ed), "não está rodando") {
		t.Errorf("painel deveria avisar que o Docker não está rodando:\n%s", frameText(ed))
	}
}

func TestDockerToggleKeybinding(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.dockerExec = fakeDockerExec(nil)
	ed.HandleEvent(input.NewResizeEvent(80, 30))

	feed := func(r rune) {
		ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: r}})
	}
	feed(' ')
	feed('d')
	feed('o')
	if ed.mode != ModeDocker || !ed.dock.open {
		t.Fatalf("Espaço d o deveria abrir o painel docker, mode=%v open=%v", ed.mode, ed.dock.open)
	}

	// de novo fecha
	feed(' ')
	feed('d')
	feed('o')
	if ed.dock.open {
		t.Fatalf("Espaço d o de novo deveria fechar o painel")
	}
}

func TestDockerActions(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.HandleEvent(input.NewResizeEvent(80, 30))

	var cmds []string
	ed.start = func(command, dir string, notify func(bool)) (*runner.Job, error) {
		cmds = append(cmds, command)
		return nil, nil
	}

	// estado montado direto, sem goroutine (determinístico)
	setup := func() {
		ed.dock = dockerState{
			open: true,
			containers: []docker.Container{
				{ID: "abc123", Name: "db", State: "running", Status: "Up 2 hours"},
				{ID: "def456", Name: "worker", State: "exited", Status: "Exited (1)"},
			},
			services: []docker.ComposeService{
				{Name: "proj-api-1", Service: "api", State: "running"},
			},
			images: []docker.Image{{ID: "sha256:aaa", Repository: "postgres", Tag: "16"}},
		}
		ed.mode = ModeDocker
	}
	feed := func(r rune) {
		ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: r}})
	}
	tab := func() { feedKey(ed, input.KeyTab) }
	last := func() string {
		if len(cmds) == 0 {
			return ""
		}
		return cmds[len(cmds)-1]
	}

	cases := []struct {
		name string
		do   func()
		want string
	}{
		{"stop container running", func() { setup(); feed('s') }, "docker stop abc123"},
		{"start container parado", func() { setup(); feed('j'); feed('s') }, "docker start def456"},
		{"restart container", func() { setup(); feed('r') }, "docker restart abc123"},
		{"logs container", func() { setup(); feed('l') }, "docker logs -f --tail 200 abc123"},
		{"stop serviço compose", func() { setup(); tab(); feed('s') }, "docker compose stop api"},
		{"restart serviço compose", func() { setup(); tab(); feed('r') }, "docker compose restart api"},
		{"rebuild serviço compose", func() { setup(); tab(); feed('b') }, "docker compose up -d --build api"},
	}
	for _, c := range cases {
		cmds = cmds[:0]
		c.do()
		if last() != c.want {
			t.Errorf("%s: comando = %q, esperava %q", c.name, last(), c.want)
		}
	}
}

func feedKey(ed *Editor, code input.KeyCode) {
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: code}})
}

func TestDockerRebuildOnlyCompose(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.HandleEvent(input.NewResizeEvent(80, 30))
	var cmds []string
	ed.start = func(command, dir string, notify func(bool)) (*runner.Job, error) {
		cmds = append(cmds, command)
		return nil, nil
	}
	ed.dock = dockerState{open: true, containers: []docker.Container{{ID: "abc", Name: "db", State: "running"}}}
	ed.mode = ModeDocker
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 'b'}})
	if len(cmds) != 0 {
		t.Errorf("rebuild num container não deveria rodar nada, rodou %v", cmds)
	}
	if !strings.Contains(ed.Frame().Status.Text(), "Compose") {
		t.Errorf("deveria avisar que rebuild é do Compose, status: %q", ed.Frame().Status.Text())
	}
}

func TestDockerRemoveConfirm(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.HandleEvent(input.NewResizeEvent(80, 30))
	var cmds []string
	ed.start = func(command, dir string, notify func(bool)) (*runner.Job, error) {
		cmds = append(cmds, command)
		return nil, nil
	}
	ed.dock = dockerState{open: true, containers: []docker.Container{{ID: "abc", Name: "db", State: "running"}}}
	ed.mode = ModeDocker

	// 'd' abre a confirmação (não roda nada ainda)
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 'd'}})
	if ed.mode != ModePrompt {
		t.Fatalf("'d' deveria abrir o prompt de confirmação, mode=%v", ed.mode)
	}
	if len(cmds) != 0 {
		t.Fatalf("nada deveria rodar antes de confirmar, rodou %v", cmds)
	}
	// responde 's' e Enter → container running usa rm -f
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 's'}})
	feedKey(ed, input.KeyEnter)
	if len(cmds) != 1 || cmds[0] != "docker rm -f abc" {
		t.Fatalf("esperava 'docker rm -f abc' após confirmar, veio %v", cmds)
	}

	// agora negando: 'n' não roda
	cmds = nil
	ed.dock = dockerState{open: true, containers: []docker.Container{{ID: "xyz", Name: "db2", State: "exited"}}}
	ed.mode = ModeDocker
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 'd'}})
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 'n'}})
	feedKey(ed, input.KeyEnter)
	if len(cmds) != 0 {
		t.Fatalf("negar não deveria remover nada, rodou %v", cmds)
	}
}

func TestDockerExecMessage(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.HandleEvent(input.NewResizeEvent(80, 30))
	var cmds []string
	ed.start = func(command, dir string, notify func(bool)) (*runner.Job, error) {
		cmds = append(cmds, command)
		return nil, nil
	}
	ed.dock = dockerState{open: true, containers: []docker.Container{{ID: "abc", Name: "db", State: "running"}}}
	ed.mode = ModeDocker
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyRune, Rune: 'x'}})
	if len(cmds) != 0 {
		t.Errorf("'x' não deveria rodar nada, rodou %v", cmds)
	}
	if !strings.Contains(ed.Frame().Status.Text(), "terminal") {
		t.Errorf("'x' deveria mandar usar o terminal, status: %q", ed.Frame().Status.Text())
	}
}

func TestDockerTabSwitch(t *testing.T) {
	ed := New(nil, "t.go")
	defer ed.Close()
	ed.dockerExec = fakeDockerExec(nil)
	ed.HandleEvent(input.NewResizeEvent(80, 30))
	ed.openDocker()
	drainOne(t, ed)

	tab := func() dockerTab { return ed.dock.tab }
	if tab() != dockContainers {
		t.Fatalf("aba inicial deveria ser Containers, veio %v", tab())
	}
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyTab}})
	if tab() != dockCompose {
		t.Fatalf("Tab deveria ir para Compose, veio %v", tab())
	}
	ed.HandleEvent(input.Event{Kind: input.EventKey, Key: input.Key{Code: input.KeyTab, Mod: input.ModShift}})
	if tab() != dockContainers {
		t.Fatalf("Shift+Tab deveria voltar para Containers, veio %v", tab())
	}
}
