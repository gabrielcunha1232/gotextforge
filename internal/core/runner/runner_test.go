package runner

import (
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCleanLine(t *testing.T) {
	tests := map[string]string{
		"texto simples":                        "texto simples",
		"\x1b[32mok\x1b[0m pacote":             "ok pacote",
		"baixando 10%\rbaixando 50%\rpronto":   "pronto",
		"linha windows\r":                      "linha windows",
		"\x1b]0;titulo\x07depois":              "depois",
		"\x1b[2K\x1b[1A\x1b[1G => [1/4] build": " => [1/4] build",
	}
	for in, want := range tests {
		if got := CleanLine(in); got != want {
			t.Errorf("CleanLine(%q) = %q, esperava %q", in, got, want)
		}
	}
}

func run(t *testing.T, command string) Snapshot {
	t.Helper()
	var notified atomic.Int32
	j, err := Start(command, t.TempDir(), func(final bool) { notified.Add(1) })
	if err != nil {
		t.Fatalf("Start(%q): %v", command, err)
	}
	if !j.Wait(60 * time.Second) {
		j.Cancel()
		t.Fatalf("comando %q nao terminou", command)
	}
	time.Sleep(20 * time.Millisecond)
	if notified.Load() == 0 {
		t.Error("o editor deveria ser notificado pelo menos no fim")
	}
	return j.Snapshot()
}

func TestComandoComSucesso(t *testing.T) {
	s := run(t, "go version")
	if s.State != Succeeded || s.ExitCode != 0 {
		t.Fatalf("estado %v codigo %d", s.State, s.ExitCode)
	}
	if len(s.Lines) == 0 || !strings.HasPrefix(s.Lines[0], "go version go") {
		t.Errorf("saida = %q", s.Lines)
	}
}

func TestComandoComVariasLinhasEStderr(t *testing.T) {
	s := run(t, "echo um&& echo dois&& go build ./nao-existe")
	if s.State != Failed || s.ExitCode == 0 {
		t.Fatalf("deveria falhar: estado %v codigo %d", s.State, s.ExitCode)
	}
	all := strings.Join(s.Lines, "\n")
	if !strings.Contains(all, "um") || !strings.Contains(all, "dois") {
		t.Errorf("saida padrao faltando: %q", all)
	}
	if len(s.Lines) < 3 || !strings.HasPrefix(s.Lines[2], "go:") {
		t.Errorf("a mensagem de erro do go (stderr) deveria aparecer depois do echo: %q", all)
	}
}

func TestComandoInexistente(t *testing.T) {
	s := run(t, "comando-que-nao-existe-xyz")
	if s.State != Failed {
		t.Errorf("comando inexistente deveria falhar: %v", s.State)
	}
}

func TestCancelar(t *testing.T) {
	long := "sleep 30"
	if runtime.GOOS == "windows" {
		long = "ping -n 30 127.0.0.1"
	}
	j, err := Start(long, t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	start := time.Now()
	j.Cancel()
	if !j.Wait(10 * time.Second) {
		t.Fatal("Cancel deveria encerrar o comando")
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("cancelar demorou %v", d)
	}
	if s := j.Snapshot(); s.State != Canceled {
		t.Errorf("estado = %v, esperava Canceled", s.State)
	}
}

func TestComandoVazio(t *testing.T) {
	if _, err := Start("   ", ".", nil); err == nil {
		t.Error("comando vazio deveria dar erro")
	}
}
