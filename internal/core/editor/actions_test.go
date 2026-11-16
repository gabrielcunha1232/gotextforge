package editor

import (
	"strings"
	"testing"
)

func TestSwitchTabSaiDoInsert(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	typeKeys(t, ed, "Ax")
	ed.SwitchTab(0)
	if ed.path != "a.go" || ed.mode != ModeNormal {
		t.Fatalf("aba %q modo %v, quero a.go em NORMAL", ed.path, ed.mode)
	}
	ed.SwitchTab(1)
	typeKeys(t, ed, "u")
	if bufferText(ed) != "package b" {
		t.Errorf("o x digitado antes do clique deveria ser um passo de undo só: %q", bufferText(ed))
	}
	ed.SwitchTab(7)
	if ed.path != "b.go" {
		t.Errorf("índice inválido não deveria trocar de aba")
	}
}

func TestCloseTabSemAlteracoes(t *testing.T) {
	ed, _ := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.CloseTab(0)
	if len(ed.docs) != 1 || ed.path != "b.go" {
		t.Fatalf("docs %d, atual %q", len(ed.docs), ed.path)
	}
	ed.CloseTab(0)
	if ed.ShouldQuit() {
		t.Fatal("fechar a última aba pelo clique não deveria sair do editor")
	}
	if len(ed.docs) != 1 || ed.path != "" || !ed.showIntro() {
		t.Errorf("deveria voltar à tela inicial: docs %d, caminho %q", len(ed.docs), ed.path)
	}
}

func TestCloseTabComAlteracoesPergunta(t *testing.T) {
	ed, saver := editorWithFiles(t, map[string]string{"a.go": "package a\n", "b.go": "package b\n"}, "a.go", "b.go")
	ed.SwitchTab(0)
	typeKeys(t, ed, "Ax<Esc>")
	ed.SwitchTab(1)

	ed.CloseTab(0)
	if ed.mode != ModePrompt || ed.path != "a.go" || !strings.Contains(ed.Frame().Status.Text(), "a.go tem alterações") {
		t.Fatalf("deveria perguntar na aba a.go: modo %v, aba %q, linha %q", ed.mode, ed.path, ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "c<CR>")
	if len(ed.docs) != 2 || !ed.buf.Modified() {
		t.Fatalf("cancelar não deveria fechar nem salvar")
	}

	ed.CloseTab(0)
	typeKeys(t, ed, "n<CR>")
	if len(ed.docs) != 1 || ed.path != "b.go" || saver.calls != 0 {
		t.Fatalf("não deveria fechar sem salvar: docs %d, aba %q, salvos %v", len(ed.docs), ed.path, saver.calls)
	}

	typeKeys(t, ed, "Ay<Esc>")
	ed.CloseTab(0)
	typeKeys(t, ed, "s<CR>")
	if saver.calls != 1 || saver.path != "b.go" || ed.path != "" {
		t.Errorf("sim deveria salvar e fechar: salvos %d (%q), aba %q", saver.calls, saver.path, ed.path)
	}
}
