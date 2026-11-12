package fuzzy

import "testing"

var files = []string{
	"cmd/gotextforge/main.go",
	"internal/core/editor/editor.go",
	"internal/core/editor/editor_test.go",
	"internal/core/editor/explorer.go",
	"internal/core/buffer/buffer.go",
	"internal/core/engine/parser.go",
	"internal/frontend/tui/renderer.go",
	"README.md",
	"go.mod",
}

func TestScoreAceitaSubsequencia(t *testing.T) {
	tests := []struct {
		text, query string
		ok          bool
	}{
		{"internal/core/editor/editor.go", "edtgo", true},
		{"internal/core/editor/editor.go", "EDITOR", true},
		{"README.md", "rdm", true},
		{"README.md", "xyz", false},
		{"go.mod", "mod.go", false},
		{"qualquer", "", true},
	}
	for _, tt := range tests {
		if _, _, ok := Score(tt.text, tt.query); ok != tt.ok {
			t.Errorf("Score(%q, %q) ok = %v, esperava %v", tt.text, tt.query, ok, tt.ok)
		}
	}
}

func TestPosicoesMarcadas(t *testing.T) {
	_, pos, ok := Score("cmd/main.go", "main")
	if !ok || len(pos) != 4 || pos[0] != 4 {
		t.Errorf("posicoes = %v", pos)
	}
}

func TestFilterOrdena(t *testing.T) {
	tests := []struct {
		query string
		first string
	}{
		{"main", "cmd/gotextforge/main.go"},
		{"editor", "internal/core/editor/editor.go"},
		{"edtest", "internal/core/editor/editor_test.go"},
		{"parser", "internal/core/engine/parser.go"},
		{"rend", "internal/frontend/tui/renderer.go"},
		{"readme", "README.md"},
		{"expl", "internal/core/editor/explorer.go"},
	}
	for _, tt := range tests {
		got := Filter(files, tt.query, 0)
		if len(got) == 0 || got[0].Text != tt.first {
			first := ""
			if len(got) > 0 {
				first = got[0].Text
			}
			t.Errorf("Filter(%q) primeiro = %q, esperava %q", tt.query, first, tt.first)
		}
	}
}

func TestFilterLimite(t *testing.T) {
	if got := Filter(files, "", 3); len(got) != 3 {
		t.Errorf("limite 3 devolveu %d", len(got))
	}
	if got := Filter(files, "zzz", 0); len(got) != 0 {
		t.Errorf("sem correspondencia deveria ser vazio: %v", got)
	}
}
