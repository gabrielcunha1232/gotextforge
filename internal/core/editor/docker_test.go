package editor

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gabrielcunha1232/gotextforge/internal/core/docker"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
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
