package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/gabrielcunha1232/gotextforge/internal/core/input"
)

type model struct {
	editor *editor.Editor
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.editor.HandleEvent(input.NewResizeEvent(msg.Width, msg.Height))

	case tea.KeyPressMsg:
		key := TranslateKey(msg)
		if key == (input.Key{Code: input.KeyRune, Rune: 'c', Mod: input.ModCtrl}) {
			return m, tea.Quit
		}
		if key != (input.Key{}) {
			m.editor.HandleEvent(input.NewKeyEvent(key))
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(RenderFrame(m.editor.Frame()))
	v.AltScreen = true
	return v
}

func Run(ed *editor.Editor) error {
	p := tea.NewProgram(model{
		editor: ed,
	})
	_, err := p.Run()
	return err
}
