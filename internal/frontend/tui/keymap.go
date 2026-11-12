package tui

import (
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TranslateKey(msg tea.KeyPressMsg) input.Key {
	var mod input.Mod
	if msg.Mod.Contains(tea.ModCtrl) {
		mod |= input.ModCtrl
	}
	if msg.Mod.Contains(tea.ModAlt) {
		mod |= input.ModAlt
	}

	switch msg.Code {
	case tea.KeyEscape:
		return input.Key{Code: input.KeyEsc, Mod: mod}
	case tea.KeyEnter:
		return input.Key{Code: input.KeyEnter, Mod: mod}
	case tea.KeyBackspace:
		return input.Key{Code: input.KeyBackspace, Mod: mod}
	case tea.KeyTab:
		return input.Key{Code: input.KeyTab, Mod: mod}
	case tea.KeyUp:
		return input.Key{Code: input.KeyUp, Mod: mod}
	case tea.KeyDown:
		return input.Key{Code: input.KeyDown, Mod: mod}
	case tea.KeyLeft:
		return input.Key{Code: input.KeyLeft, Mod: mod}
	case tea.KeyRight:
		return input.Key{Code: input.KeyRight, Mod: mod}
	case tea.KeyHome:
		return input.Key{Code: input.KeyHome, Mod: mod}
	case tea.KeyEnd:
		return input.Key{Code: input.KeyEnd, Mod: mod}
	case tea.KeyPgUp:
		return input.Key{Code: input.KeyPageUp, Mod: mod}
	case tea.KeyPgDown:
		return input.Key{Code: input.KeyPageDown, Mod: mod}
	case tea.KeyDelete:
		return input.Key{Code: input.KeyDelete, Mod: mod}
	case tea.KeySpace:
		return input.Key{Code: input.KeyRune, Rune: ' ', Mod: mod}
	}

	var r rune
	if msg.Text != "" {
		r, _ = utf8.DecodeRuneInString(msg.Text)
	} else if msg.Code > 0 && msg.Code <= unicode.MaxRune {
		r = msg.Code
	}

	if r != 0 {
		if mod&input.ModCtrl != 0 {
			r = unicode.ToLower(r)
		} else if msg.Mod.Contains(tea.ModShift) && msg.Text == "" {
			r = unicode.ToUpper(r)
		}
		return input.Key{Code: input.KeyRune, Rune: r, Mod: mod}
	}

	return input.Key{}
}
