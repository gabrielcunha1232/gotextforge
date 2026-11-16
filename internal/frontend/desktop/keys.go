package desktop

import (
	"unicode"
	"unicode/utf8"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

type KeyInput struct {
	Key   string `json:"key"`
	Ctrl  bool   `json:"ctrl,omitempty"`
	Alt   bool   `json:"alt,omitempty"`
	Shift bool   `json:"shift,omitempty"`
}

var namedKeys = map[string]input.KeyCode{
	"Escape":     input.KeyEsc,
	"Enter":      input.KeyEnter,
	"Backspace":  input.KeyBackspace,
	"Tab":        input.KeyTab,
	"ArrowUp":    input.KeyUp,
	"ArrowDown":  input.KeyDown,
	"ArrowLeft":  input.KeyLeft,
	"ArrowRight": input.KeyRight,
	"Home":       input.KeyHome,
	"End":        input.KeyEnd,
	"PageUp":     input.KeyPageUp,
	"PageDown":   input.KeyPageDown,
	"Delete":     input.KeyDelete,
}

func translateKey(k KeyInput) (input.Key, bool) {
	var mod input.Mod
	if k.Ctrl {
		mod |= input.ModCtrl
	}
	if k.Alt {
		mod |= input.ModAlt
	}

	if code, ok := namedKeys[k.Key]; ok {
		if k.Shift {
			mod |= input.ModShift
		}
		return input.Key{Code: code, Mod: mod}, true
	}

	r, size := utf8.DecodeRuneInString(k.Key)
	if r == utf8.RuneError || size != len(k.Key) || unicode.IsControl(r) {
		return input.Key{}, false
	}
	if k.Ctrl {
		r = unicode.ToLower(r)
	}
	return input.Key{Code: input.KeyRune, Rune: r, Mod: mod}, true
}
