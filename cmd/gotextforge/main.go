package main

import (
	"fmt"
	"os"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/frontend/tui"
)

func main() {
	ed := editor.New()
	if err := tui.Run(ed); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao executar editor: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("gotextforge v0")
}
