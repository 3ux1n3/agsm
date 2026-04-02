package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Search  key.Binding
	Refresh key.Binding
	Rename  key.Binding
	Delete  key.Binding
	New     key.Binding
	Enter   key.Binding
	Quit    key.Binding
	Esc     key.Binding
	Tab     key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up", "move up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down", "move down")),
		Search:  key.NewBinding(key.WithKeys("/", "ctrl+f"), key.WithHelp("/", "search")),
		Refresh: key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "refresh")),
		Rename:  key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "rename")),
		Delete:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		New:     key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "new")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "resume")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Esc:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	}
}
