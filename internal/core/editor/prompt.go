package editor

import "github.com/gabrielcunha1232/gotextforge/internal/core/input"

type prompt struct {
	label  string
	text   []rune
	back   Mode
	submit func(text string)
}

func (e *Editor) openPrompt(label, initial string, submit func(string)) {
	back := e.mode
	if back == ModePrompt {
		back = ModeNormal
	}
	e.prompt = &prompt{label: label, text: []rune(initial), back: back, submit: submit}
	e.mode = ModePrompt
}

func (e *Editor) handlePromptKey(k input.Key) {
	p := e.prompt
	if p == nil {
		e.mode = ModeNormal
		return
	}
	switch {
	case k.Code == input.KeyEsc || isCtrl(k, 'c'):
		e.mode = p.back
		e.prompt = nil
	case k.Code == input.KeyEnter:
		e.mode = p.back
		e.prompt = nil
		p.submit(string(p.text))
	case k.Code == input.KeyBackspace:
		if len(p.text) > 0 {
			p.text = p.text[:len(p.text)-1]
		}
	case isCtrl(k, 'u'):
		p.text = p.text[:0]
	case k.Code == input.KeyRune && k.Mod == 0 && k.Rune != 0:
		p.text = append(p.text, k.Rune)
	}
}
