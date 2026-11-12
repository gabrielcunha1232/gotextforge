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
	opened := false
	for _, arg := range os.Args[1:] {
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			ed.SetRoot(arg)
			ed.ShowExplorer()
			continue
		}
		ed.OpenFile(arg)
		opened = true
	}
	if opened {
		ed.SwitchToFirst()
	}

	err := tui.Run(ed)
	ed.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao executar editor: %v\n", err)
		os.Exit(1)
	}
}
