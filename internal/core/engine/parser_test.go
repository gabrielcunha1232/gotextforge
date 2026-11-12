package engine

import (
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func feedAll(p *Parser, keys string) (Command, bool) {
	var cmd Command
	var done bool
	for _, r := range keys {
		cmd, done = p.Feed(input.Key{Code: input.KeyRune, Rune: r})
	}
	return cmd, done
}

func TestParser(t *testing.T) {
	tests := []struct {
		keys string
		want Command
	}{
		{"j", Command{Motion: Motion{Kind: MotionDown}}},
		{"5j", Command{Count: 5, Motion: Motion{Kind: MotionDown}}},
		{"12w", Command{Count: 12, Motion: Motion{Kind: MotionWordStart}}},
		{"0", Command{Motion: Motion{Kind: MotionLineStart}}},
		{"10j", Command{Count: 10, Motion: Motion{Kind: MotionDown}}},
		{"gg", Command{Motion: Motion{Kind: MotionFileStart}}},
		{"5gg", Command{Count: 5, Motion: Motion{Kind: MotionFileStart}}},
		{"G", Command{Motion: Motion{Kind: MotionFileEnd}}},
		{"fx", Command{Motion: Motion{Kind: MotionFindForward, Char: 'x'}}},
		{"2t)", Command{Count: 2, Motion: Motion{Kind: MotionTillForward, Char: ')'}}},
		{"F(", Command{Motion: Motion{Kind: MotionFindBackward, Char: '('}}},
		{"dw", Command{Operator: OpDelete, Motion: Motion{Kind: MotionWordStart}}},
		{"dd", Command{Operator: OpDelete, Motion: Motion{Kind: MotionCurrentLine}}},
		{"3dd", Command{Count: 3, Operator: OpDelete, Motion: Motion{Kind: MotionCurrentLine}}},
		{"2d3w", Command{Count: 6, Operator: OpDelete, Motion: Motion{Kind: MotionWordStart}}},
		{"d0", Command{Operator: OpDelete, Motion: Motion{Kind: MotionLineStart}}},
		{"cc", Command{Operator: OpChange, Motion: Motion{Kind: MotionCurrentLine}}},
		{"yy", Command{Operator: OpYank, Motion: Motion{Kind: MotionCurrentLine}}},
		{"ciw", Command{Operator: OpChange, Motion: Motion{Kind: MotionTextObject, Object: ObjectWord, Inner: true}}},
		{"daw", Command{Operator: OpDelete, Motion: Motion{Kind: MotionTextObject, Object: ObjectWord}}},
		{"dap", Command{Operator: OpDelete, Motion: Motion{Kind: MotionTextObject, Object: ObjectParagraph}}},
		{"ci\"", Command{Operator: OpChange, Motion: Motion{Kind: MotionTextObject, Object: ObjectDoubleQuote, Inner: true}}},
		{"di(", Command{Operator: OpDelete, Motion: Motion{Kind: MotionTextObject, Object: ObjectParen, Inner: true}}},
		{"dib", Command{Operator: OpDelete, Motion: Motion{Kind: MotionTextObject, Object: ObjectParen, Inner: true}}},
		{"ya{", Command{Operator: OpYank, Motion: Motion{Kind: MotionTextObject, Object: ObjectBrace}}},
		{"dgg", Command{Operator: OpDelete, Motion: Motion{Kind: MotionFileStart}}},
		{"dfx", Command{Operator: OpDelete, Motion: Motion{Kind: MotionFindForward, Char: 'x'}}},
		{"x", Command{Operator: OpDelete, Motion: Motion{Kind: MotionRight}}},
		{"3x", Command{Count: 3, Operator: OpDelete, Motion: Motion{Kind: MotionRight}}},
		{"X", Command{Operator: OpDelete, Motion: Motion{Kind: MotionLeft}}},
		{"D", Command{Operator: OpDelete, Motion: Motion{Kind: MotionLineEnd}}},
		{"C", Command{Operator: OpChange, Motion: Motion{Kind: MotionLineEnd}}},
		{"S", Command{Operator: OpChange, Motion: Motion{Kind: MotionCurrentLine}}},
		{"Y", Command{Operator: OpYank, Motion: Motion{Kind: MotionCurrentLine}}},
		{"i", Command{Action: ActionInsert}},
		{"3p", Command{Count: 3, Action: ActionPasteAfter}},
		{"rx", Command{Action: ActionReplace, Char: 'x'}},
		{"4J", Command{Count: 4, Action: ActionJoin}},
		{":", Command{Action: ActionCommandLine}},
	}

	for _, tt := range tests {
		t.Run(tt.keys, func(t *testing.T) {
			var p Parser
			got, done := feedAll(&p, tt.keys)
			if !done {
				t.Fatalf("comando %q nao terminou (pendente: %q)", tt.keys, p.Pending())
			}
			if got != tt.want {
				t.Errorf("Feed(%q) = %+v\nesperava       %+v", tt.keys, got, tt.want)
			}
			if p.Pending() != "" {
				t.Errorf("depois de terminar, Pending() deveria estar vazio, veio %q", p.Pending())
			}
		})
	}
}

func TestParserPendente(t *testing.T) {
	for _, keys := range []string{"d", "2", "2d", "d2", "g", "f", "dt", "ci", "da", "r", "y3"} {
		var p Parser
		if _, done := feedAll(&p, keys); done {
			t.Errorf("%q nao deveria terminar um comando", keys)
		}
		if p.Pending() != keys {
			t.Errorf("Pending() = %q, esperava %q", p.Pending(), keys)
		}
	}
}

func TestParserCancela(t *testing.T) {
	for _, keys := range []string{"dz", "gz", "diz", "dx", "dp", "cI"} {
		var p Parser
		if _, done := feedAll(&p, keys); done {
			t.Errorf("%q deveria ser cancelado, nao virar comando", keys)
		}
		if p.Pending() != "" {
			t.Errorf("%q: depois de cancelar Pending() = %q", keys, p.Pending())
		}
	}
}

func TestParserEscCancela(t *testing.T) {
	var p Parser
	feedAll(&p, "2d")
	if _, done := p.Feed(input.Key{Code: input.KeyEsc}); done {
		t.Fatal("Esc nao deveria gerar comando")
	}
	if got, done := feedAll(&p, "w"); !done || got.Operator != OpNone || got.Count != 0 {
		t.Errorf("depois do Esc, w deveria ser um movimento simples; veio %+v", got)
	}
}

func TestParserTeclasEspeciais(t *testing.T) {
	tests := []struct {
		key  input.Key
		want Command
	}{
		{input.Key{Code: input.KeyDown}, Command{Motion: Motion{Kind: MotionDown}}},
		{input.Key{Code: input.KeyPageDown}, Command{Motion: Motion{Kind: MotionPageDown}}},
		{input.Key{Code: input.KeyDelete}, Command{Operator: OpDelete, Motion: Motion{Kind: MotionRight}}},
		{input.Key{Code: input.KeyRune, Rune: 'r', Mod: input.ModCtrl}, Command{Action: ActionRedo}},
		{input.Key{Code: input.KeyRune, Rune: 'c', Mod: input.ModCtrl}, Command{Action: ActionQuit}},
	}
	for _, tt := range tests {
		var p Parser
		got, done := p.Feed(tt.key)
		if !done || got != tt.want {
			t.Errorf("Feed(%v) = %+v (done=%v), esperava %+v", tt.key, got, done, tt.want)
		}
	}

	var p Parser
	feedAll(&p, "d")
	if got, done := p.Feed(input.Key{Code: input.KeyDown}); !done || got.Operator != OpDelete || got.Motion.Kind != MotionDown {
		t.Errorf("d + seta para baixo deveria ser dj; veio %+v", got)
	}
}

func TestParserCountGigante(t *testing.T) {
	var p Parser
	got, _ := feedAll(&p, "99999999999j")
	if got.Count != countLimit {
		t.Errorf("count = %d, esperava o limite %d", got.Count, countLimit)
	}
	got, _ = feedAll(&p, "999d999w")
	if got.Count != countLimit {
		t.Errorf("count multiplicado = %d, esperava o limite %d", got.Count, countLimit)
	}
}
