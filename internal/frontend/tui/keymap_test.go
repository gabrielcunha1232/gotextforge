package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TestTranslateKey(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want input.Key
	}{
		{
			name: "letra minuscula d",
			msg:  tea.KeyPressMsg{Code: 'd', Text: "d"},
			want: input.Key{Code: input.KeyRune, Rune: 'd'},
		},
		{
			name: "letra maiuscula A",
			msg:  tea.KeyPressMsg{Code: 'a', Text: "A", Mod: tea.ModShift},
			want: input.Key{Code: input.KeyRune, Rune: 'A'},
		},
		{
			name: "caractere acentuado cedilha",
			msg:  tea.KeyPressMsg{Code: 'ç', Text: "ç"},
			want: input.Key{Code: input.KeyRune, Rune: 'ç'},
		},
		{
			name: "tecla Esc",
			msg:  tea.KeyPressMsg{Code: tea.KeyEscape},
			want: input.Key{Code: input.KeyEsc},
		},
		{
			name: "tecla Enter",
			msg:  tea.KeyPressMsg{Code: tea.KeyEnter},
			want: input.Key{Code: input.KeyEnter},
		},
		{
			name: "Ctrl+a",
			msg:  tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl},
			want: input.Key{Code: input.KeyRune, Rune: 'a', Mod: input.ModCtrl},
		},
		{
			name: "Alt+x",
			msg:  tea.KeyPressMsg{Code: 'x', Mod: tea.ModAlt},
			want: input.Key{Code: input.KeyRune, Rune: 'x', Mod: input.ModAlt},
		},
		{
			name: "Ctrl+Alt+Up",
			msg:  tea.KeyPressMsg{Code: tea.KeyUp, Mod: tea.ModCtrl | tea.ModAlt},
			want: input.Key{Code: input.KeyUp, Mod: input.ModCtrl | input.ModAlt},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateKey(tt.msg)
			if got != tt.want {
				t.Errorf("TranslateKey(%+v) = %+v (%q), esperava %+v (%q)",
					tt.msg, got, got.String(), tt.want, tt.want.String())
			}
		})
	}
}
