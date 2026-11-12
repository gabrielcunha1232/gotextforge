package editor

import (
	"testing"

	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

func TestEditorHandleEventUpdatesStatus(t *testing.T) {
	ed := New()

	if got := ed.Frame().Status; got != "última tecla: (nenhuma)" {
		t.Fatalf("status inicial inesperado: %q", got)
	}

	ed.HandleEvent(input.NewKeyEvent(input.Key{Rune: 'a', Mod: input.ModCtrl}))

	want := "última tecla: <C-a>"
	if got := ed.Frame().Status; got != want {
		t.Errorf("Frame().Status = %q, esperava %q", got, want)
	}
}
