package tui

import "github.com/charmbracelet/bubbles/textinput"

func newDirectoryInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Directory: "
	input.Placeholder = "/path/to/project"
	input.CharLimit = 400
	return input
}

func newSessionNameInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Session Name: "
	input.Placeholder = "optional"
	input.CharLimit = 120
	return input
}

func newPromptInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Initial Prompt: "
	input.Placeholder = "optional"
	input.CharLimit = 400
	return input
}
