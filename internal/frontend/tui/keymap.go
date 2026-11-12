package tui

import (
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TranslateKey(msg tea.KeyPressMsg) input.Key {
	var baseMod input.Mod
	if msg.Mod.Contains(tea.ModCtrl) {
		baseMod |= input.ModCtrl
	}
	if msg.Mod.Contains(tea.ModAlt) {
		baseMod |= input.ModAlt
	}

	specialMod := baseMod
	if msg.Mod.Contains(tea.ModShift) {
		specialMod |= input.ModShift
	}

	switch msg.Code {
	case tea.KeyEscape:
		return input.Key{Code: input.KeyEsc, Mod: specialMod}
	case tea.KeyEnter:
		return input.Key{Code: input.KeyEnter, Mod: specialMod}
	case tea.KeyBackspace:
		return input.Key{Code: input.KeyBackspace, Mod: specialMod}
	case tea.KeyTab:
		return input.Key{Code: input.KeyTab, Mod: specialMod}
	case tea.KeyUp:
		return input.Key{Code: input.KeyUp, Mod: specialMod}
	case tea.KeyDown:
		return input.Key{Code: input.KeyDown, Mod: specialMod}
	case tea.KeyLeft:
		return input.Key{Code: input.KeyLeft, Mod: specialMod}
	case tea.KeyRight:
		return input.Key{Code: input.KeyRight, Mod: specialMod}
	case tea.KeyHome:
		return input.Key{Code: input.KeyHome, Mod: specialMod}
	case tea.KeyEnd:
		return input.Key{Code: input.KeyEnd, Mod: specialMod}
	case tea.KeyPgUp:
		return input.Key{Code: input.KeyPageUp, Mod: specialMod}
	case tea.KeyPgDown:
		return input.Key{Code: input.KeyPageDown, Mod: specialMod}
	case tea.KeyDelete:
		return input.Key{Code: input.KeyDelete, Mod: specialMod}
	case tea.KeySpace:
		return input.Key{Code: input.KeyRune, Rune: ' ', Mod: baseMod}
	}

	var r rune
	if msg.Text != "" {
		r, _ = utf8.DecodeRuneInString(msg.Text)
	} else if msg.Code > 0 && msg.Code <= unicode.MaxRune {
		r = msg.Code
	}

	if r != 0 {
		if baseMod&input.ModCtrl != 0 {
			r = unicode.ToLower(r)
		} else if msg.Mod.Contains(tea.ModShift) && msg.Text == "" {
			r = unicode.ToUpper(r)
		}
		return input.Key{Code: input.KeyRune, Rune: r, Mod: baseMod}
	}

	return input.Key{}
}
