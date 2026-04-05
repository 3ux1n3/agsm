package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func newDirectoryInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Directory: "
	input.Placeholder = "/path/to/project"
	input.CharLimit = 400
	styleInput(&input)
	return input
}

func newSessionNameInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Session Name: "
	input.Placeholder = "optional"
	input.CharLimit = 120
	styleInput(&input)
	return input
}

func newPromptInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Initial Prompt: "
	input.Placeholder = "optional"
	input.CharLimit = 400
	styleInput(&input)
	return input
}

func styleInput(input *textinput.Model) {
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "14", Light: "4"}).Bold(true)
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "15", Light: "0"})
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "8", Light: "8"})
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Dark: "13", Light: "5"})
}
