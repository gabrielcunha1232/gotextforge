package docker

import (
	"errors"
	"testing"
)

const sampleContainers = `{"Command":"\"docker-entrypoint.s…\"","CreatedAt":"2026-11-13 10:00:00 +0000 UTC","ID":"abc123","Image":"postgres:16","Names":"db","Ports":"0.0.0.0:5432->5432/tcp","State":"running","Status":"Up 2 hours"}
{"ID":"not a valid json line
{"Command":"\"/worker\"","CreatedAt":"2026-11-13 09:00:00 +0000 UTC","ID":"def456","Image":"myapp/worker:latest","Names":"worker,worker_old","Ports":"","State":"exited","Status":"Exited (1) 3 minutes ago"}
`

func fakeExec(out string, err error) Exec {
	return func(args ...string) ([]byte, error) {
		return []byte(out), err
	}
}

func TestListContainers(t *testing.T) {
	c := New(fakeExec(sampleContainers, nil))
	got, err := c.ListContainers()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("esperava 2 containers (a linha ruim é pulada), veio %d: %+v", len(got), got)
	}

	db := got[0]
	if db.ID != "abc123" || db.Name != "db" || db.State != "running" {
		t.Errorf("container running incorreto: %+v", db)
	}
	if db.Ports != "0.0.0.0:5432->5432/tcp" {
		t.Errorf("portas incorretas: %q", db.Ports)
	}

	worker := got[1]
	if worker.State != "exited" {
		t.Errorf("esperava State exited, veio %q", worker.State)
	}
	if worker.Status != "Exited (1) 3 minutes ago" {
		t.Errorf("Status do container com erro incorreto: %q", worker.Status)
	}
	if worker.Name != "worker" {
		t.Errorf("esperava só o primeiro nome (worker), veio %q", worker.Name)
	}
}

func TestListContainersExecError(t *testing.T) {
	want := errors.New("falhou")
	c := New(fakeExec("", want))
	_, err := c.ListContainers()
	if !errors.Is(err, want) {
		t.Fatalf("esperava o erro do Exec repassado, veio %v", err)
	}
}

func TestListContainersEmpty(t *testing.T) {
	c := New(fakeExec("", nil))
	got, err := c.ListContainers()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("esperava lista vazia, veio %+v", got)
	}
}

const sampleImages = `{"Containers":"N/A","CreatedAt":"2026-11-01 12:00:00 +0000 UTC","CreatedSince":"2 weeks ago","Digest":"<none>","ID":"sha256:aaa","Repository":"postgres","SharedSize":"N/A","Size":"438MB","Tag":"16","UniqueSize":"N/A","VirtualSize":"438MB"}
linha de lixo
{"Containers":"N/A","CreatedAt":"2026-11-10 08:00:00 +0000 UTC","CreatedSince":"3 days ago","Digest":"<none>","ID":"sha256:bbb","Repository":"myapp/worker","SharedSize":"N/A","Size":"22.5MB","Tag":"latest","UniqueSize":"N/A","VirtualSize":"22.5MB"}
`

func TestListImages(t *testing.T) {
	c := New(fakeExec(sampleImages, nil))
	got, err := c.ListImages()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("esperava 2 imagens (linha ruim pulada), veio %d: %+v", len(got), got)
	}
	if got[0].Repository != "postgres" || got[0].Tag != "16" || got[0].Size != "438MB" {
		t.Errorf("imagem 0 incorreta: %+v", got[0])
	}
	if got[1].ID != "sha256:bbb" || got[1].Repository != "myapp/worker" {
		t.Errorf("imagem 1 incorreta: %+v", got[1])
	}
}

func TestListImagesExecError(t *testing.T) {
	want := errors.New("falhou")
	c := New(fakeExec("", want))
	_, err := c.ListImages()
	if !errors.Is(err, want) {
		t.Fatalf("esperava o erro do Exec repassado, veio %v", err)
	}
}

func TestInfoOK(t *testing.T) {
	c := New(fakeExec("29.9.1\n", nil))
	ver, err := c.Info()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ver != "29.9.1" {
		t.Errorf("esperava versão aparada, veio %q", ver)
	}
}

func TestInfoDaemonDown(t *testing.T) {
	c := New(fakeExec("", errors.New("failed to connect to the docker API")))
	_, err := c.Info()
	if !errors.Is(err, ErrDaemonDown) {
		t.Fatalf("esperava ErrDaemonDown, veio %v", err)
	}
}

// NDJSON: um objeto por linha (compose v2 e algumas versões do v5).
const sampleComposeNDJSON = `{"Name":"proj-api-1","Service":"api","State":"running","Status":"Up 2 hours","Health":"healthy","ExitCode":0,"Publishers":[{"URL":"0.0.0.0","TargetPort":8080,"PublishedPort":8080,"Protocol":"tcp"}]}
{"Name":"proj-worker-1","Service":"worker","State":"exited","Status":"Exited (1) 3 minutes ago","Health":"","ExitCode":1,"Publishers":null}
`

// Array: a mesma informação como um único array JSON (outras versões do v5).
const sampleComposeArray = `[
  {"Name":"proj-api-1","Service":"api","State":"running","Status":"Up 2 hours","Health":"healthy","ExitCode":0,"Publishers":[{"URL":"0.0.0.0","TargetPort":8080,"PublishedPort":8080,"Protocol":"tcp"}]},
  {"Name":"proj-worker-1","Service":"worker","State":"exited","Status":"Exited (1) 3 minutes ago","Health":"","ExitCode":1,"Publishers":null}
]`

func checkComposeSample(t *testing.T, got []ComposeService) {
	t.Helper()
	if len(got) != 2 {
		t.Fatalf("esperava 2 serviços, veio %d: %+v", len(got), got)
	}
	api := got[0]
	if api.Service != "api" || api.State != "running" || api.Health != "healthy" {
		t.Errorf("serviço api incorreto: %+v", api)
	}
	if api.Ports != "0.0.0.0:8080->8080/tcp" {
		t.Errorf("portas do api incorretas: %q", api.Ports)
	}
	worker := got[1]
	if worker.State != "exited" || worker.ExitCode != 1 {
		t.Errorf("serviço worker (saiu com erro) incorreto: %+v", worker)
	}
	if worker.Ports != "" {
		t.Errorf("worker sem publishers deveria ter portas vazias, veio %q", worker.Ports)
	}
}

func TestComposePSNDJSON(t *testing.T) {
	c := New(fakeExec(sampleComposeNDJSON, nil))
	got, err := c.ComposePS()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	checkComposeSample(t, got)
}

func TestComposePSArray(t *testing.T) {
	c := New(fakeExec(sampleComposeArray, nil))
	got, err := c.ComposePS()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	checkComposeSample(t, got)
}

func TestComposePSExecError(t *testing.T) {
	want := errors.New("no configuration file provided")
	c := New(fakeExec("", want))
	_, err := c.ComposePS()
	if !errors.Is(err, want) {
		t.Fatalf("esperava o erro do Exec repassado, veio %v", err)
	}
}
