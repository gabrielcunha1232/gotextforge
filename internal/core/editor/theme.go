package editor

import (
	"math"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
	"github.com/gabrielcunha1232/gotextforge/internal/core/syntax"
)

type Theme struct {
	Name        string
	PaintBg     bool
	Bg          render.Color
	OnAccent    render.Color
	OnDark      render.Color
	BgHighlight render.Color
	BgStatus    render.Color
	BgFloat     render.Color
	BgSel       render.Color
	SearchBg    render.Color
	Fg          render.Color
	FgDark      render.Color
	Gutter      render.Color
	Dark3       render.Color
	Comment     render.Color
	Blue        render.Color
	Blue1       render.Color
	Magenta     render.Color
	Orange      render.Color
	Yellow      render.Color
	Green       render.Color
	Red         render.Color
	Teal        render.Color
	Border      render.Color
	Logo        [3]render.Color
	Syntax      [syntax.KindFunc + 1]render.Style
	Modes       map[Mode]render.Color
}

var ThemeGo = newThemeGo()

func newThemeGo() Theme {
	t := Theme{
		Name:        "go",
		Bg:          render.RGB(0x0b, 0x0e, 0x13),
		OnAccent:    render.RGB(0x0b, 0x0e, 0x13),
		BgHighlight: render.RGB(0x1a, 0x22, 0x2e),
		BgStatus:    render.RGB(0x16, 0x1b, 0x22),
		BgFloat:     render.RGB(0x0e, 0x11, 0x16),
		BgSel:       render.RGB(0x20, 0x30, 0x40),
		SearchBg:    render.RGB(0x3d, 0x59, 0xa1),
		Fg:          render.RGB(0xd6, 0xe2, 0xec),
		FgDark:      render.RGB(0xaa, 0xb7, 0xc4),
		Gutter:      render.RGB(0x3a, 0x46, 0x54),
		Dark3:       render.RGB(0x5a, 0x6b, 0x7a),
		Comment:     render.RGB(0x5f, 0x70, 0x83),
		Blue:        render.RGB(0x00, 0xad, 0xd8),
		Blue1:       render.RGB(0x38, 0xbd, 0xd0),
		Magenta:     render.RGB(0xb0, 0x8c, 0xf0),
		Orange:      render.RGB(0xe8, 0xa2, 0x4a),
		Yellow:      render.RGB(0xe6, 0xc0, 0x72),
		Green:       render.RGB(0x5b, 0xc9, 0x8a),
		Red:         render.RGB(0xff, 0x6b, 0x7a),
		Teal:        render.RGB(0x2b, 0xb3, 0xa3),
		Border:      render.RGB(0x2b, 0x7a, 0x99),
	}
	t.Logo = [3]render.Color{t.Blue, t.Teal, t.Green}
	t.Syntax = [...]render.Style{
		syntax.KindPlain:       {},
		syntax.KindKeyword:     {Fg: t.Magenta, Italic: true},
		syntax.KindString:      {Fg: t.Green},
		syntax.KindNumber:      {Fg: t.Orange},
		syntax.KindComment:     {Fg: t.Comment, Italic: true},
		syntax.KindPredeclared: {Fg: t.Blue1},
		syntax.KindFunc:        {Fg: t.Blue},
	}
	t.Modes = map[Mode]render.Color{
		ModeNormal:   t.Blue,
		ModeInsert:   t.Green,
		ModeCommand:  t.Yellow,
		ModeSearch:   t.Yellow,
		ModePrompt:   t.Yellow,
		ModeExplorer: t.Teal,
		ModeFinder:   t.Magenta,
		ModeTerminal: t.Green,
		ModeDocker:   t.Blue1,
	}
	return t
}

func (t *Theme) lineNr() render.Style { return render.Style{Fg: t.Gutter} }

func (t *Theme) cursorLineNr() render.Style {
	return render.Style{Fg: t.Orange, Bold: true, Bg: t.BgHighlight}
}

func (t *Theme) tilde() render.Style   { return render.Style{Fg: t.Dark3} }
func (t *Theme) statusB() render.Style { return render.Style{Fg: t.Fg, Bg: t.Gutter} }
func (t *Theme) statusC() render.Style { return render.Style{Fg: t.FgDark, Bg: t.BgStatus} }
func (t *Theme) statusError() render.Style {
	return render.Style{Fg: t.Red, Bg: t.BgStatus, Bold: true}
}
func (t *Theme) introSub() render.Style  { return render.Style{Fg: t.Comment, Italic: true} }
func (t *Theme) introKey() render.Style  { return render.Style{Fg: t.Blue, Bold: true} }
func (t *Theme) introDesc() render.Style { return render.Style{Fg: t.FgDark} }
func (t *Theme) search() render.Style    { return render.Style{Fg: t.Fg, Bg: t.SearchBg} }
func (t *Theme) statusPending() render.Style {
	return render.Style{Fg: t.Orange, Bg: t.BgStatus, Bold: true}
}

func (t *Theme) searchCurrent() render.Style {
	return render.Style{Fg: t.on(t.Orange), Bg: t.Orange, Bold: true}
}

var modeNames = map[Mode]string{
	ModeNormal:   "NORMAL",
	ModeInsert:   "INSERT",
	ModeCommand:  "COMMAND",
	ModeSearch:   "SEARCH",
	ModePrompt:   "PROMPT",
	ModeExplorer: "ARQUIVOS",
	ModeFinder:   "BUSCAR",
	ModeTerminal: "TERMINAL",
	ModeDocker:   "DOCKER",
}

func withBg(s render.Style, bg render.Color) render.Style {
	s.Bg = bg
	return s
}

func (t *Theme) Palette() map[string]render.Color {
	p := map[string]render.Color{
		"bg":           t.Bg,
		"on-accent":    t.OnAccent,
		"bg-highlight": t.BgHighlight,
		"bg-status":    t.BgStatus,
		"bg-float":     t.BgFloat,
		"bg-sel":       t.BgSel,
		"fg":           t.Fg,
		"fg-dark":      t.FgDark,
		"gutter":       t.Gutter,
		"dark3":        t.Dark3,
		"comment":      t.Comment,
		"blue":         t.Blue,
		"blue1":        t.Blue1,
		"magenta":      t.Magenta,
		"orange":       t.Orange,
		"yellow":       t.Yellow,
		"green":        t.Green,
		"red":          t.Red,
		"teal":         t.Teal,
		"border":       t.Border,
	}
	for m, c := range t.Modes {
		p["mode-"+strings.ToLower(modeNames[m])] = c
	}
	return p
}

var ThemeRetro = newThemeRetro()

func newThemeRetro() Theme {
	ink := render.RGB(0x2b, 0x21, 0x18)
	orange := render.RGB(0xe8, 0x64, 0x3c)
	mustard := render.RGB(0xf2, 0xb1, 0x34)
	teal := render.RGB(0x2a, 0x9d, 0x8f)
	t := Theme{
		Name:        "retro",
		PaintBg:     true,
		Bg:          render.RGB(0xf1, 0xe4, 0xc8),
		OnAccent:    ink,
		OnDark:      render.RGB(0xfb, 0xf3, 0xde),
		BgHighlight: render.RGB(0xea, 0xdb, 0xba),
		BgStatus:    render.RGB(0xfb, 0xf3, 0xde),
		BgFloat:     render.RGB(0xfb, 0xf3, 0xde),
		BgSel:       render.RGB(0xf6, 0xd5, 0x8a),
		SearchBg:    mustard,
		Fg:          ink,
		FgDark:      render.RGB(0x5a, 0x4a, 0x3a),
		Gutter:      render.RGB(0xb3, 0x9f, 0x7e),
		Dark3:       render.RGB(0x86, 0x72, 0x5a),
		Comment:     render.RGB(0x6e, 0x5e, 0x48),
		Blue:        render.RGB(0x1f, 0x5f, 0x6f),
		Blue1:       render.RGB(0x16, 0x69, 0x60),
		Magenta:     render.RGB(0x8c, 0x4a, 0x7a),
		Orange:      render.RGB(0xa8, 0x42, 0x1e),
		Yellow:      render.RGB(0x85, 0x5a, 0x08),
		Green:       render.RGB(0x42, 0x66, 0x1f),
		Red:         render.RGB(0xad, 0x2e, 0x25),
		Teal:        teal,
		Border:      ink,
		Logo:        [3]render.Color{orange, mustard, teal},
	}
	t.Syntax = [...]render.Style{
		syntax.KindPlain:       {},
		syntax.KindKeyword:     {Fg: t.Red, Bold: true},
		syntax.KindString:      {Fg: t.Green},
		syntax.KindNumber:      {Fg: t.Magenta},
		syntax.KindComment:     {Fg: t.Comment, Italic: true},
		syntax.KindPredeclared: {Fg: t.Blue1},
		syntax.KindFunc:        {Fg: t.Blue, Bold: true},
	}
	t.Modes = map[Mode]render.Color{
		ModeNormal:   teal,
		ModeInsert:   orange,
		ModeCommand:  mustard,
		ModeSearch:   mustard,
		ModePrompt:   mustard,
		ModeExplorer: teal,
		ModeFinder:   orange,
		ModeTerminal: mustard,
		ModeDocker:   teal,
	}
	return t
}

var themes = []*Theme{&ThemeGo, &ThemeRetro}

func ThemeNames() []string {
	names := make([]string, len(themes))
	for i, t := range themes {
		names[i] = t.Name
	}
	return names
}

func (e *Editor) SetTheme(name string) bool {
	for _, t := range themes {
		if t.Name == name {
			e.theme = t
			return true
		}
	}
	return false
}

func (t *Theme) on(bg render.Color) render.Color {
	if t.OnDark == (render.Color{}) || contrast(t.OnAccent, bg) >= contrast(t.OnDark, bg) {
		return t.OnAccent
	}
	return t.OnDark
}

func luminance(c render.Color) float64 {
	ch := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(c.R) + 0.7152*ch(c.G) + 0.0722*ch(c.B)
}

func contrast(a, b render.Color) float64 {
	la, lb := luminance(a), luminance(b)
	return (max(la, lb) + 0.05) / (min(la, lb) + 0.05)
}
