package syntax

import (
	"testing"
)

type lines []string

func (l lines) LineCount() int    { return len(l) }
func (l lines) Line(i int) []rune { return []rune(l[i]) }

func TestCacheStateAt(t *testing.T) {
	src := lines{
		"package main",
		"/* comeca",
		"meio",
		"fim */",
		"var q = `select",
		"from t",
		"`",
		"func main() {}",
	}

	want := []State{
		StateNormal,
		StateNormal,
		StateBlockComment,
		StateBlockComment,
		StateNormal,
		StateRawString,
		StateRawString,
		StateNormal,
	}

	var c Cache
	for i := len(want) - 1; i >= 0; i-- {
		if got := c.StateAt(src, i); got != want[i] {
			t.Errorf("StateAt(%d) = %v, esperava %v", i, got, want[i])
		}
	}
}

func TestCacheInvalidate(t *testing.T) {
	src := lines{"x := 1", "y := 2", "z := 3"}
	var c Cache
	if got := c.StateAt(src, 2); got != StateNormal {
		t.Fatalf("StateAt(2) = %v", got)
	}

	src[0] = "x := 1 /* abre"
	if got := c.StateAt(src, 2); got != StateNormal {
		t.Fatalf("sem invalidar, o cache deveria continuar devolvendo o estado antigo; veio %v", got)
	}

	c.Invalidate(0)
	if got := c.StateAt(src, 1); got != StateBlockComment {
		t.Errorf("depois de invalidar, StateAt(1) = %v, esperava StateBlockComment", got)
	}
	if got := c.StateAt(src, 2); got != StateBlockComment {
		t.Errorf("depois de invalidar, StateAt(2) = %v, esperava StateBlockComment", got)
	}
}

func TestCacheLinhaAlemDoFim(t *testing.T) {
	src := lines{"a", "b"}
	var c Cache
	if got := c.StateAt(src, 99); got != StateNormal {
		t.Errorf("StateAt(99) = %v", got)
	}
}

func TestNextStateConcordaComHighlightLine(t *testing.T) {
	lineList := []string{
		"",
		"package main",
		"x := 1 // comentario com * e /",
		"a / b * c",
		"/* abre",
		"fecha */ x",
		"*/",
		"/* a */ b /* c",
		"s := \"/* nao e comentario\"",
		"s := \"`\"",
		"q := `abre",
		"fecha` + `outra",
		"`",
		"c := '`'",
		"// /* comentario de linha esconde o bloco",
	}
	for _, line := range lineList {
		for _, in := range []State{StateNormal, StateBlockComment, StateRawString} {
			_, want := HighlightLine([]rune(line), in)
			if got := NextState([]rune(line), in); got != want {
				t.Errorf("linha %q estado %v: NextState = %v, HighlightLine diz %v", line, in, got, want)
			}
		}
	}
}
