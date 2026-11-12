package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type text string

func (s text) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(s))
	return int64(n), err
}

type failAfter struct {
	written string
}

func (f failAfter) WriteTo(w io.Writer) (int64, error) {
	n, _ := io.WriteString(w, f.written)
	return int64(n), errors.New("disco cheio")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("arquivo temporario esquecido: %s", e.Name())
		}
	}
}

func TestSaveCriaArquivoNovo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "novo.go")

	if err := Save(path, text("package main\n")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := readFile(t, path); got != "package main\n" {
		t.Errorf("conteudo = %q", got)
	}
	assertNoTempFiles(t, dir)
}

func TestSaveSubstituiConteudoInteiro(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("conteudo antigo bem maior que o novo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Save(path, text("novo\n")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := readFile(t, path); got != "novo\n" {
		t.Errorf("conteudo = %q, esperava %q (sem sobras do antigo)", got, "novo\n")
	}
	assertNoTempFiles(t, dir)
}

func TestSaveComErroPreservaOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	const original = "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Save(path, failAfter{written: "pack"})
	if err == nil {
		t.Fatal("esperava erro, veio nil")
	}
	if got := readFile(t, path); got != original {
		t.Errorf("original foi corrompido: %q", got)
	}
	assertNoTempFiles(t, dir)
}

func TestSavePastaInexistente(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nao", "existe", "a.go")
	if err := Save(path, text("x")); err == nil {
		t.Fatal("esperava erro ao salvar em pasta inexistente")
	}
}

func TestSavePreservaPermissoes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permissoes unix nao se aplicam no Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(path, []byte("echo oi\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, text("echo tchau\n")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("permissoes = %v, esperava 0755", info.Mode().Perm())
	}
}

func TestSaveSegueSymlink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.go")
	link := filepath.Join(dir, "link.go")
	if err := os.WriteFile(real, []byte("antigo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink indisponivel neste sistema: %v", err)
	}

	if err := Save(link, text("novo\n")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := readFile(t, real); got != "novo\n" {
		t.Errorf("arquivo real = %q, esperava %q", got, "novo\n")
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("o symlink foi substituido por um arquivo comum")
	}
}
