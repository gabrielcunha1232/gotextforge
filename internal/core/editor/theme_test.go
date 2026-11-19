package editor

import (
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

func runEx(ed *Editor, cmd string) {
	press(ed, ':')
	for _, r := range cmd {
		press(ed, r)
	}
	pressCode(ed, input.KeyEnter)
}

func TestColorschemeTrocaOTema(t *testing.T) {
	ed := editorFrom(t, "package main\n\nfunc main() {}\n")
	if ed.Theme().Name != "go" {
		t.Fatalf("tema inicial = %q, esperava go", ed.Theme().Name)
	}
	runEx(ed, "colorscheme retro")
	if ed.Theme() != &ThemeRetro {
		t.Fatalf("tema depois do :colorscheme retro = %q", ed.Theme().Name)
	}
	if got := ed.Frame().Lines[0][0].Style; got != ThemeRetro.cursorLineNr() {
		t.Errorf("numero da linha atual com estilo %+v, esperava o do retro", got)
	}
	runEx(ed, "colo go")
	if ed.Theme() != &ThemeGo {
		t.Errorf("tema depois do :colo go = %q", ed.Theme().Name)
	}
}

func TestColorschemeSemNomeMostraOAtual(t *testing.T) {
	ed := editorFrom(t, "x\n")
	runEx(ed, "colorscheme")
	if ed.message != "go" {
		t.Errorf("mensagem = %q, esperava go", ed.message)
	}
}

func TestColorschemeDesconhecidoDaErro(t *testing.T) {
	ed := editorFrom(t, "x\n")
	runEx(ed, "colorscheme neon")
	if !isErrorMessage(ed.message) || !strings.Contains(ed.message, "neon") {
		t.Errorf("mensagem = %q, esperava erro citando neon", ed.message)
	}
	if ed.Theme() != &ThemeGo {
		t.Errorf("tema mudou para %q depois de um nome invalido", ed.Theme().Name)
	}
}

func TestTemasCompletos(t *testing.T) {
	want := ThemeGo.Palette()
	for _, th := range themes {
		p := th.Palette()
		for name := range want {
			if _, ok := p[name]; !ok {
				t.Errorf("tema %s sem a cor %s", th.Name, name)
			}
		}
		for m := range modeNames {
			if _, ok := th.Modes[m]; !ok {
				t.Errorf("tema %s sem cor para o modo %s", th.Name, modeNames[m])
			}
		}
	}
}

func TestRetroTextoLegivelNoCreme(t *testing.T) {
	th := &ThemeRetro
	texts := map[string]render.Color{
		"fg": th.Fg, "fg-dark": th.FgDark, "comment": th.Comment, "blue": th.Blue, "blue1": th.Blue1,
		"magenta": th.Magenta, "orange": th.Orange, "yellow": th.Yellow, "green": th.Green, "red": th.Red,
	}
	for name, c := range texts {
		for _, bg := range []render.Color{th.Bg, th.BgHighlight} {
			if r := contrast(c, bg); r < 4.0 {
				t.Errorf("%s sobre %+v: contraste %.2f, minimo 4.0", name, bg, r)
			}
		}
	}
	for m, c := range th.Modes {
		if r := contrast(th.on(c), c); r < 4.0 {
			t.Errorf("texto do modo %s: contraste %.2f, minimo 4.0", modeNames[m], r)
		}
	}
}

func TestRetroTextoSobreFundosColoridos(t *testing.T) {
	th := &ThemeRetro
	for _, bg := range []render.Color{th.Orange, th.Blue1, th.Yellow, th.Green, th.Red, th.Gutter, th.Dark3} {
		if r := contrast(th.on(bg), bg); r < 4.0 {
			t.Errorf("texto sobre %+v: contraste %.2f, minimo 4.0", bg, r)
		}
	}
}

func TestTemaGoUsaSempreATintaEscuraNosRotulos(t *testing.T) {
	for _, bg := range []render.Color{ThemeGo.Orange, ThemeGo.Gutter, ThemeGo.Red} {
		if got := ThemeGo.on(bg); got != ThemeGo.OnAccent {
			t.Errorf("tema go trocou a cor do texto sobre %+v", bg)
		}
	}
}
