package editor

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
	"github.com/gabrielcunha1232/gotextforge/internal/core/render"
)

type fakeSaver struct {
	calls   int
	path    string
	content string
	err     error
}

func (f *fakeSaver) save(path string, src io.WriterTo) error {
	f.calls++
	f.path = path
	if f.err != nil {
		return f.err
	}
	var sb strings.Builder
	if _, err := src.WriteTo(&sb); err != nil {
		return err
	}
	f.content = sb.String()
	return nil
}

func editorWithSaver(t *testing.T, src string) (*Editor, *fakeSaver) {
	t.Helper()
	ed := editorFrom(t, src)
	ed.path = "/projeto/teste.go"
	fs := &fakeSaver{}
	ed.save = fs.save
	return ed, fs
}

func pressCtrl(ed *Editor, r rune) {
	ed.HandleEvent(input.NewKeyEvent(input.Key{Code: input.KeyRune, Rune: r, Mod: input.ModCtrl}))
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		keys        string
		saveErr     error
		wantQuit    bool
		wantSaves   int
		wantContent string
		wantMessage string
		wantDirty   bool
	}{
		{name: ":w salva", src: "abc\n", keys: "AX<Esc>:w<CR>", wantSaves: 1, wantContent: "abcX\n", wantMessage: `"teste.go" 1L gravado`},
		{name: ":w sem alteracoes salva mesmo assim", src: "abc\n", keys: ":w<CR>", wantSaves: 1, wantContent: "abc\n"},
		{name: ":q sem alteracoes sai", src: "abc\n", keys: ":q<CR>", wantQuit: true},
		{name: ":q com alteracoes avisa e nao sai", src: "abc\n", keys: "x:q<CR>", wantMessage: "E37", wantDirty: true},
		{name: ":q! sai descartando", src: "abc\n", keys: "x:q!<CR>", wantQuit: true, wantDirty: true},
		{name: ":wq salva e sai", src: "abc\n", keys: "x:wq<CR>", wantQuit: true, wantSaves: 1, wantContent: "bc\n"},
		{name: ":x com alteracoes salva e sai", src: "abc\n", keys: "x:x<CR>", wantQuit: true, wantSaves: 1, wantContent: "bc\n"},
		{name: ":x sem alteracoes sai sem salvar", src: "abc\n", keys: ":x<CR>", wantQuit: true},
		{name: ":wq com erro ao salvar nao sai", src: "abc\n", keys: "x:wq<CR>", saveErr: errors.New("disco cheio"), wantSaves: 1, wantMessage: "erro ao salvar: disco cheio", wantDirty: true},
		{name: "espacos em volta do comando", src: "abc\n", keys: ":  q  <CR>", wantQuit: true},
		{name: "comando desconhecido", src: "abc\n", keys: ":foo<CR>", wantMessage: "E492: comando desconhecido: foo"},
		{name: "Esc cancela o comando", src: "abc\n", keys: ":q<Esc>", wantQuit: false},
		{name: "Backspace corrige o comando", src: "abc\n", keys: ":wx<BS><CR>", wantSaves: 1, wantContent: "abc\n"},
		{name: "Backspace com linha vazia sai do modo comando", src: "abc\n", keys: ":<BS>x", wantDirty: true},
		{name: "Enter vazio nao faz nada", src: "abc\n", keys: ":<CR>"},
		{name: "arquivo com UTF-8 invalido nao salva", src: "a\xffb\n", keys: "x:w<CR>", wantMessage: "UTF-8 inválidos", wantDirty: true},
		{name: "CRLF preservado ao salvar", src: "a\r\nb\r\n", keys: "AX<Esc>:w<CR>", wantSaves: 1, wantContent: "aX\r\nb\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed, fs := editorWithSaver(t, tt.src)
			fs.err = tt.saveErr
			typeKeys(t, ed, tt.keys)

			if ed.ShouldQuit() != tt.wantQuit {
				t.Errorf("ShouldQuit() = %v, esperava %v", ed.ShouldQuit(), tt.wantQuit)
			}
			if fs.calls != tt.wantSaves {
				t.Errorf("salvou %d vezes, esperava %d", fs.calls, tt.wantSaves)
			}
			if tt.wantSaves > 0 && tt.saveErr == nil {
				if fs.content != tt.wantContent {
					t.Errorf("conteudo salvo = %q, esperava %q", fs.content, tt.wantContent)
				}
				if fs.path != "/projeto/teste.go" {
					t.Errorf("salvou em %q, esperava o caminho completo", fs.path)
				}
			}
			if !strings.Contains(ed.Frame().Status.Text(), tt.wantMessage) {
				t.Errorf("status = %q, esperava conter %q", ed.Frame().Status.Text(), tt.wantMessage)
			}
			if ed.buf.Modified() != tt.wantDirty {
				t.Errorf("Modified() = %v, esperava %v", ed.buf.Modified(), tt.wantDirty)
			}
			if ed.Mode() != ModeNormal {
				t.Errorf("depois do comando o modo deveria ser normal, veio %v", ed.Mode())
			}
		})
	}
}

func TestCtrlC(t *testing.T) {
	t.Run("sem alteracoes sai", func(t *testing.T) {
		ed, _ := editorWithSaver(t, "abc\n")
		pressCtrl(ed, 'c')
		if !ed.ShouldQuit() {
			t.Error("Ctrl+c sem alteracoes deveria sair")
		}
	})

	t.Run("com alteracoes avisa e nao sai", func(t *testing.T) {
		ed, _ := editorWithSaver(t, "abc\n")
		typeKeys(t, ed, "x")
		pressCtrl(ed, 'c')
		if ed.ShouldQuit() {
			t.Error("Ctrl+c com alteracoes nao deveria sair")
		}
		if !strings.Contains(ed.Frame().Status.Text(), "E37") {
			t.Errorf("status deveria avisar E37, veio %q", ed.Frame().Status.Text())
		}
	})

	t.Run("no insert funciona como Esc", func(t *testing.T) {
		ed, _ := editorWithSaver(t, "abc\n")
		typeKeys(t, ed, "AX")
		pressCtrl(ed, 'c')
		if ed.Mode() != ModeNormal || ed.ShouldQuit() {
			t.Errorf("Ctrl+c no insert deveria voltar ao normal sem sair (modo %v, quit %v)", ed.Mode(), ed.ShouldQuit())
		}
	})

	t.Run("no modo comando cancela", func(t *testing.T) {
		ed, _ := editorWithSaver(t, "abc\n")
		typeKeys(t, ed, ":q")
		pressCtrl(ed, 'c')
		if ed.Mode() != ModeNormal || ed.ShouldQuit() {
			t.Errorf("Ctrl+c no comando deveria cancelar sem sair (modo %v, quit %v)", ed.Mode(), ed.ShouldQuit())
		}
	})
}

func TestCommandModeFrame(t *testing.T) {
	ed, _ := editorWithSaver(t, "abc\n")
	typeKeys(t, ed, ":wq")

	frame := ed.Frame()
	if frame.Status.Text() != ":wq" {
		t.Errorf("status = %q, esperava %q", frame.Status.Text(), ":wq")
	}
	if frame.Cursor != (render.ScreenPos{Row: 23, Col: 3}) {
		t.Errorf("cursor = %+v, esperava {Row:23 Col:3} (na linha de status, depois do texto)", frame.Cursor)
	}
	if frame.CursorShape != render.CursorBar {
		t.Errorf("cursor no modo comando deveria ser barra")
	}
}

func TestMensagemSomeNaProximaTecla(t *testing.T) {
	ed, _ := editorWithSaver(t, "abc\ndef\n")
	typeKeys(t, ed, ":w<CR>")
	if !strings.Contains(ed.Frame().Status.Text(), "gravado") {
		t.Fatalf("esperava mensagem de gravado, veio %q", ed.Frame().Status.Text())
	}
	typeKeys(t, ed, "j")
	if strings.Contains(ed.Frame().Status.Text(), "gravado") {
		t.Errorf("a mensagem deveria sumir depois da proxima tecla, veio %q", ed.Frame().Status.Text())
	}
	if strings.Contains(ed.Frame().Status.Text(), "[+]") {
		t.Errorf("depois de salvar nao deveria ter [+], veio %q", ed.Frame().Status.Text())
	}
}

func TestSemNomeNaoSalva(t *testing.T) {
	ed := New(nil, "")
	fs := &fakeSaver{}
	ed.save = fs.save
	typeKeys(t, ed, "iX<Esc>:w<CR>")
	if fs.calls != 0 {
		t.Error("nao deveria tentar salvar um buffer sem nome")
	}
	if !strings.Contains(ed.Frame().Status.Text(), "E32") {
		t.Errorf("status deveria avisar E32, veio %q", ed.Frame().Status.Text())
	}
}

func TestSalvarDePontaAPonta(t *testing.T) {
	const original = "\xef\xbb\xbfpackage main\r\n\r\nfunc main() {\r\n}\r\n"
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	buf, err := buffer.Load(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}

	ed := New(buf, path)
	ed.HandleEvent(input.NewResizeEvent(80, 24))
	typeKeys(t, ed, "jjoprintln(1)<Esc>:wq<CR>")

	if !ed.ShouldQuit() {
		t.Fatalf("deveria ter saido; status: %q", ed.Frame().Status.Text())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const want = "\xef\xbb\xbfpackage main\r\n\r\nfunc main() {\r\n\tprintln(1)\r\n}\r\n"
	if string(got) != want {
		t.Errorf("arquivo gravado = %q\nesperava          %q", got, want)
	}
}
