package buffer

import "testing"

func TestUndoRedoOperacaoSolta(t *testing.T) {
	b := mustLoad(t, "abc")
	b.Insert(Pos{Line: 0, Col: 1}, "X")
	b.Delete(Pos{Line: 0, Col: 0}, Pos{Line: 0, Col: 1})

	if got := allText(b); got != "Xbc" {
		t.Fatalf("texto = %q, esperava %q", got, "Xbc")
	}

	steps := []struct {
		op   func() (Pos, bool)
		want string
		ok   bool
	}{
		{op: b.Undo, want: "aXbc", ok: true},
		{op: b.Undo, want: "abc", ok: true},
		{op: b.Undo, want: "abc", ok: false},
		{op: b.Redo, want: "aXbc", ok: true},
		{op: b.Redo, want: "Xbc", ok: true},
		{op: b.Redo, want: "Xbc", ok: false},
	}
	for i, s := range steps {
		_, ok := s.op()
		if ok != s.ok {
			t.Errorf("passo %d: ok = %v, esperava %v", i, ok, s.ok)
		}
		if got := allText(b); got != s.want {
			t.Errorf("passo %d: texto = %q, esperava %q", i, got, s.want)
		}
	}
}

func TestUndoAgrupaMudancas(t *testing.T) {
	b := mustLoad(t, "fim")
	b.BeginChange(Pos{Line: 0, Col: 0})
	p := Pos{}
	for _, r := range "hello" {
		p = b.Insert(p, string(r))
	}
	b.Insert(p, "\n")
	b.EndChange()

	if got := allText(b); got != "hello\nfim" {
		t.Fatalf("texto = %q", got)
	}

	cursor, ok := b.Undo()
	if !ok {
		t.Fatal("Undo deveria funcionar")
	}
	if got := allText(b); got != "fim" {
		t.Errorf("um Undo deveria desfazer o grupo inteiro; texto = %q", got)
	}
	if cursor != (Pos{Line: 0, Col: 0}) {
		t.Errorf("cursor depois do Undo = %+v, esperava a posicao de BeginChange", cursor)
	}

	if _, ok := b.Undo(); ok {
		t.Error("nao deveria haver mais nada para desfazer")
	}

	b.Redo()
	if got := allText(b); got != "hello\nfim" {
		t.Errorf("Redo deveria refazer o grupo inteiro; texto = %q", got)
	}
}

func TestGrupoVazioNaoEntraNoHistorico(t *testing.T) {
	b := mustLoad(t, "abc")
	b.BeginChange(Pos{})
	b.EndChange()
	if _, ok := b.Undo(); ok {
		t.Error("grupo sem mudancas nao deveria gerar passo de undo")
	}
}

func TestNovaMudancaApagaRedo(t *testing.T) {
	b := mustLoad(t, "abc")
	b.Insert(Pos{Line: 0, Col: 3}, "1")
	b.Undo()
	b.Insert(Pos{Line: 0, Col: 3}, "2")

	if _, ok := b.Redo(); ok {
		t.Error("depois de uma mudanca nova, o redo antigo deveria ter sumido")
	}
	if got := allText(b); got != "abc2" {
		t.Errorf("texto = %q, esperava %q", got, "abc2")
	}
}

func TestUndoMultilinhaEJuntarLinhas(t *testing.T) {
	const src = "package main\n\nfunc main() {\n\tprintln(1)\n}"
	b := mustLoad(t, src)

	b.Delete(Pos{Line: 1, Col: 0}, Pos{Line: 3, Col: 5})
	b.Insert(Pos{Line: 0, Col: 7}, " nova\nlinha\n")
	b.Delete(Pos{Line: 0, Col: 12}, Pos{Line: 1, Col: 0})

	for {
		if _, ok := b.Undo(); !ok {
			break
		}
	}
	if got := allText(b); got != src {
		t.Errorf("depois de desfazer tudo:\n%q\nesperava o original\n%q", got, src)
	}
}

func TestModifiedAcompanhaUndo(t *testing.T) {
	b := mustLoad(t, "abc")

	b.Insert(Pos{Line: 0, Col: 0}, "X")
	if !b.Modified() {
		t.Fatal("depois de editar deveria estar modificado")
	}
	b.Undo()
	if b.Modified() {
		t.Error("desfazer tudo volta ao estado original: nao deveria estar modificado")
	}
	b.Redo()
	if !b.Modified() {
		t.Error("refazer volta a ter alteracoes")
	}

	b.MarkSaved()
	if b.Modified() {
		t.Fatal("logo depois de salvar nao deveria estar modificado")
	}
	b.Undo()
	if !b.Modified() {
		t.Error("desfazer algo que ja foi salvo deixa o buffer diferente do disco")
	}
	b.Redo()
	if b.Modified() {
		t.Error("refazer volta exatamente ao estado salvo")
	}
}
