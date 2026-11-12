package main

import (
	"fmt"
	"os"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/lsp"
	"github.com/gabrielcunha1232/gotextforge/internal/frontend/tui"
)

func main() {
	ed := editor.New(nil, "")
	ed.SetLanguageServer(func(path string) editor.LanguageClient {
		return lsp.Start(path, lsp.GoplsStarter)
	})
	for _, path := range os.Args[1:] {
		ed.OpenFile(path)
	}
	ed.SwitchToFirst()

	err := tui.Run(ed)
	ed.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao executar editor: %v\n", err)
		os.Exit(1)
	}
}
