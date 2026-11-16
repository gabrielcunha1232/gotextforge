package desktop

import (
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TestTranslateKey(t *testing.T) {
	cases := []struct {
		in   KeyInput
		want string
	}{
		{KeyInput{Key: "j"}, "j"},
		{KeyInput{Key: "A", Shift: true}, "A"},
		{KeyInput{Key: "ç"}, "ç"},
		{KeyInput{Key: "á"}, "á"},
		{KeyInput{Key: "{", Shift: true}, "{"},
		{KeyInput{Key: " "}, "<Space>"},
		{KeyInput{Key: " ", Ctrl: true}, "<C-Space>"},
		{KeyInput{Key: "r", Ctrl: true}, "<C-r>"},
		{KeyInput{Key: "R", Ctrl: true, Shift: true}, "<C-r>"},
		{KeyInput{Key: "j", Alt: true}, "<M-j>"},
		{KeyInput{Key: "Escape"}, "<Esc>"},
		{KeyInput{Key: "Enter"}, "<CR>"},
		{KeyInput{Key: "Tab", Shift: true}, "<S-Tab>"},
		{KeyInput{Key: "ArrowUp", Ctrl: true}, "<C-Up>"},
		{KeyInput{Key: "Backspace"}, "<BS>"},
		{KeyInput{Key: "Delete"}, "<Del>"},
		{KeyInput{Key: "PageDown"}, "<PageDown>"},
	}
	for _, c := range cases {
		got, ok := translateKey(c.in)
		if !ok {
			t.Errorf("%+v: não traduziu", c.in)
			continue
		}
		if got.String() != c.want {
			t.Errorf("%+v = %q, quero %q", c.in, got.String(), c.want)
		}
	}
}

func TestTranslateKeyIgnora(t *testing.T) {
	for _, k := range []string{"", "Shift", "Control", "Alt", "Dead", "F5", "Unidentified", "\x1b"} {
		if got, ok := translateKey(KeyInput{Key: k}); ok {
			t.Errorf("%q devia ser ignorada, virou %v", k, got)
		}
	}
}

func TestTranslateKeyNaoMarcaShiftEmRune(t *testing.T) {
	got, _ := translateKey(KeyInput{Key: "?", Shift: true})
	if got != (input.Key{Code: input.KeyRune, Rune: '?'}) {
		t.Errorf("got %+v", got)
	}
}
