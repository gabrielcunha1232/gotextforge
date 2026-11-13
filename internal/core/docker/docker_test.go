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
