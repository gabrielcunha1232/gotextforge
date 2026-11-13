package editor

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/runner"
)

func TestSequenciasAleatoriasNaoQuebramOEditor(t *testing.T) {
	sources := []string{
		"",
		"x",
		"package main\n\nfunc main() {\n\tfmt.Println(\"ação\", 42) // ok\n}\n",
		"a b c\n\n\nd e f\n(x [y {z}])\n\"q\" 'r' `s`\n",
		"\t\t\n  \n\n",
		"/* comentario\nque continua */\nvar q = `raw\nstring`\n",
	}
	runes := []rune("hjklwbeWBE0^$GgfFtT;{}()[]\"'`xXdDcCsSyYpPuJrioaAIO:qQ123 \tzZ.,/?nN*#HLe")
	specials := []input.KeyCode{input.KeyEsc, input.KeyEnter, input.KeyBackspace, input.KeyDelete,
		input.KeyUp, input.KeyDown, input.KeyLeft, input.KeyRight, input.KeyPageDown, input.KeyPageUp, input.KeyHome, input.KeyEnd}

	sandbox := t.TempDir()
	rng := rand.New(rand.NewPCG(42, 2026))
	for round := 0; round < 400; round++ {
		src := sources[rng.IntN(len(sources))]
		ed := editorFrom(t, src)
		ed.path = "aleatorio.go"
		ed.save = (&fakeSaver{}).save
		ed.load = fakeFiles(map[string]string{"outro.go": src, "e": "x\n"})
		ed.SetRoot(sandbox)
		ed.start = func(string, string, func(bool)) (*runner.Job, error) {
			return nil, errors.New("comandos desativados no teste aleatorio")
		}
		ed.dockerExec = nil
		if rng.IntN(2) == 0 {
			ed.OpenFile("outro.go")
		}
		width, height := 5+rng.IntN(120), 2+rng.IntN(40)
		ed.HandleEvent(input.NewResizeEvent(width, height))

		var history []string
		for step := 0; step < 200; step++ {
			var k input.Key
			switch n := rng.IntN(20); {
			case n == 0:
				k = input.Key{Code: specials[rng.IntN(len(specials))]}
			case n == 1:
				k = input.Key{Code: input.KeyRune, Rune: 'r', Mod: input.ModCtrl}
			default:
				k = input.Key{Code: input.KeyRune, Rune: runes[rng.IntN(len(runes))]}
			}
			history = append(history, k.String())

			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic %v\narquivo %q tela %dx%d\nteclas %v", r, src, width, height, history)
					}
				}()
				ed.HandleEvent(input.NewKeyEvent(k))
				f := ed.Frame()

				c := ed.Cursor()
				if c.Line < 0 || c.Line >= ed.buf.LineCount() || c.Col < 0 || c.Col > ed.buf.LineLen(c.Line) {
					t.Fatalf("cursor invalido %+v (linhas=%d)\nteclas %v", c, ed.buf.LineCount(), history)
				}
				for i, l := range f.Lines {
					if l.Width() > width {
						t.Fatalf("linha %d com largura %d > %d\nteclas %v", i, l.Width(), width, history)
					}
				}
				if f.Status.Width() > width {
					t.Fatalf("status com largura %d > %d\nteclas %v", f.Status.Width(), width, history)
				}
			}()

			if ed.ShouldQuit() {
				break
			}
		}
	}
}
