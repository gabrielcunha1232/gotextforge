package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gabrielcunha1232/gotextforge/internal/core/buffer"
	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/frontend/tui"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "uso: gotextforge [arquivo]")
		os.Exit(1)
	}

	var buf *buffer.Buffer
	path := ""
	if len(os.Args) == 2 {
		path = os.Args[1]
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro ao abrir arquivo %q: %v\n", path, err)
			os.Exit(1)
		}
		buf, err = buffer.Load(f)
		f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro ao ler arquivo %q: %v\n", path, err)
			os.Exit(1)
		}
	}

	ed := editor.New(buf, path)
	if strings.HasSuffix(path, ".go") {
		ed.SetLanguageClient(lsp.Start(path, lsp.GoplsStarter))
	}
	err := tui.Run(ed)
	ed.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao executar editor: %v\n", err)
		os.Exit(1)
	}
}
