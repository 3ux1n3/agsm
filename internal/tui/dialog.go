package tui

import "github.com/charmbracelet/bubbles/textinput"

func newRenameInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Rename: "
	input.CharLimit = 120
	return input
}
