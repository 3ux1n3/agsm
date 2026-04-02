package tui

import "github.com/charmbracelet/bubbles/textinput"

func newSearchInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "title, folder, agent"
	input.CharLimit = 256
	return input
}
